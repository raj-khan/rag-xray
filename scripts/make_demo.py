"""Turn the frames from capture.mjs into docs/media/demo.webp.

Identical consecutive frames are merged, and long pauses (such as waiting
for a slow local model) are capped, so the animation stays short.
"""
import json
import sys
from pathlib import Path

from PIL import Image, ImageChops

out = Path(sys.argv[1] if len(sys.argv) > 1 else "docs/media")
frame_dir = out / "frames"
times = json.loads((frame_dir / "times.json").read_text())
paths = sorted(frame_dir.glob("*.jpg"))[: len(times)]

WIDTH = 960
MAX_HOLD_MS = 1600

frames, durations = [], []
for i, p in enumerate(paths):
    img = Image.open(p).convert("RGB")
    img = img.resize((WIDTH, round(img.height * WIDTH / img.width)), Image.LANCZOS)
    ms = (times[i + 1] - times[i]) if i + 1 < len(times) else 1500
    if frames and not ImageChops.difference(frames[-1], img).getbbox():
        durations[-1] += ms
        continue
    frames.append(img)
    durations.append(ms)

durations = [min(max(d, 60), MAX_HOLD_MS) for d in durations]
durations[-1] = 2500
frames[0].save(out / "demo.webp", save_all=True, append_images=frames[1:],
               duration=durations, loop=0, quality=70, method=6)
print(f"{len(paths)} frames -> {len(frames)} unique, {sum(durations) / 1000:.1f}s, "
      f"{(out / 'demo.webp').stat().st_size / 1e6:.1f} MB")
