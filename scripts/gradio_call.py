#!/usr/bin/env python3
"""Gradio SenseVoice — stdlib only, real SSE streaming.
"""
import sys, json, http.client, re, time, html as _html

HOST = "localhost:7860"
TAG_RE = re.compile(r"<[^>]+>")
WS_RE = re.compile(r"\s+")

def extract(html_text):
    t = TAG_RE.sub("", html_text)
    t = _html.unescape(t)
    return WS_RE.sub(" ", t).strip()

def emit(obj):
    print(json.dumps(obj, ensure_ascii=False))
    sys.stdout.flush()

def main():
    if len(sys.argv) < 2:
        emit({"error": "usage: gradio_call.py <audio_path>"})
        sys.exit(1)

    audio = sys.argv[1]
    session = f"go_{int(time.time()*1000)}"

    try:
        # Step 1: Submit job (short timeout)
        body = json.dumps({
            "data": [{"path": audio}, "file", "auto", 3, True],
            "event_data": None, "trigger_id": None,
            "session_hash": session,
        })
        c = http.client.HTTPConnection(HOST, timeout=10)
        c.request("POST", "/gradio_api/call/transcribe", body=body,
                  headers={"Content-Type": "application/json"})
        cr = c.getresponse()
        cb = cr.read().decode()
        c.close()
        if cr.status != 200:
            emit({"error": f"Gradio HTTP {cr.status}: {cb[:200]}"})
            sys.exit(1)

        emit({"text": "🔊 正在转写..."})

        # Step 2: Connect to queue SSE (long timeout for long-running job)
        q = http.client.HTTPConnection(HOST, timeout=300)
        q.request("GET", f"/gradio_api/queue/data?session_hash={session}")
        qr = q.getresponse()
        if qr.status != 200:
            emit({"error": f"queue HTTP {qr.status}"})
            sys.exit(1)

        # Step 3: Read SSE events
        buf = b""
        seen = set()

        while True:
            chunk = qr.read(1)
            if not chunk:
                break
            buf += chunk
            if buf.endswith(b"\n\n"):
                for line in buf.decode().strip().split("\n"):
                    line = line.strip()
                    if not line.startswith("data: "):
                        continue
                    try:
                        msg = json.loads(line[6:])
                    except json.JSONDecodeError:
                        continue
                    mt = msg.get("msg", "")
                    if mt == "unexpected_error":
                        emit({"error": msg.get("message", "unknown")})
                        q.close()
                        sys.exit(1)
                    if mt in ("process_generating", "process_completed", "complete"):
                        arr = msg.get("output", {}).get("data") or msg.get("data")
                        if isinstance(arr, list):
                            for item in arr:
                                if isinstance(item, str):
                                    t = extract(item)
                                    if t and t not in seen:
                                        seen.add(t)
                                        emit({"text": t})
                    if mt in ("process_completed", "complete"):
                        emit({"done": True, "full_text": "\n\n".join(seen)})
                        q.close()
                        return
                buf = b""

    except Exception as e:
        emit({"error": str(e)})
        sys.exit(1)
    finally:
        try: q.close()
        except: pass

if __name__ == "__main__":
    main()
