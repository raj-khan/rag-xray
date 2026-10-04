"""Turn the frames from capture.mjs into docs/media/demo.webp and demo.mp4.

Identical consecutive frames are merged, and long pauses (such as waiting
for a slow local model) are capped, so the demo stays short.

The MP4 needs ffmpeg: either on PATH, or `pip install imageio-ffmpeg`.
"""
import json
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

from PIL import Image, ImageChops

out = Path(sys.argv[1] if len(sys.argv) > 1 else "docs/media")
frame_dir = out / "frames"
times = json.loads((frame_dir / "times.json").read_text())
paths = sorted(frame_dir.glob("*.jpg"))[: len(times)]

GIF_WIDTH = 960
MAX_HOLD_MS = 1600

frames, durations = [], []
for i, p in enumerate(paths):
    img = Image.open(p).convert("RGB")
    ms = (times[i + 1] - times[i]) if i + 1 < len(times) else 1500
    if frames and not ImageChops.difference(frames[-1], img).getbbox():
        durations[-1] += ms
        continue
    frames.append(img)
    durations.append(ms)

durations = [min(max(d, 60), MAX_HOLD_MS) for d in durations]
durations[-1] = 2500
print(f"{len(paths)} frames -> {len(frames)} unique, {sum(durations) / 1000:.1f}s")

# Animated WebP: plays inline everywhere, including GitHub READMEs.
small = [f.resize((GIF_WIDTH, round(f.height * GIF_WIDTH / f.width)), Image.LANCZOS) for f in frames]
small[0].save(out / "demo.webp", save_all=True, append_images=small[1:],
              duration=durations, loop=0, quality=70, method=6)
print(f"demo.webp {(out / 'demo.webp').stat().st_size / 1e6:.1f} MB")


def find_ffmpeg():
    if shutil.which("ffmpeg"):
        return "ffmpeg"
    try:
        import imageio_ffmpeg
        return imageio_ffmpeg.get_ffmpeg_exe()
    except ImportError:
        return None


ffmpeg = find_ffmpeg()
if not ffmpeg:
    print("ffmpeg not found: skipping demo.mp4 (pip install imageio-ffmpeg)")
    sys.exit(0)

# MP4 (H.264): full resolution, plays in browsers and on GitHub's file view.
with tempfile.TemporaryDirectory() as tmp:
    lines = []
    for i, (f, d) in enumerate(zip(frames, durations)):
        name = f"{i:05d}.png"
        f.save(Path(tmp) / name)
        lines += [f"file '{name}'", f"duration {d / 1000:.3f}"]
    lines.append(f"file '{len(frames) - 1:05d}.png'")  # concat demuxer needs the last frame twice
    (Path(tmp) / "list.txt").write_text("\n".join(lines))
    subprocess.run([
        ffmpeg, "-y", "-loglevel", "error", "-f", "concat", "-safe", "0", "-i", "list.txt",
        "-vf", "fps=30,format=yuv420p", "-c:v", "libx264", "-crf", "26", "-preset", "slow",
        "-movflags", "+faststart", str((out / "demo.mp4").resolve()),
    ], cwd=tmp, check=True)
print(f"demo.mp4 {(out / 'demo.mp4').stat().st_size / 1e6:.1f} MB")
