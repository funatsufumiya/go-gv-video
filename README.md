## go-gv-video

GV codec ([Extreme Gpu Friendly Video Format](https://github.com/Ushio/ofxExtremeGpuVideo)) video decoder for Go (pure Go).

Ported from [rust-gv-video](https://github.com/funatsufumiya/rust-gv-video)

This lib only provides file (or memory) loader, if you need player, please use [ebiten_gvvideo](https://github.com/funatsufumiya/ebiten_gvvideo) or create your one.

> [!WARNING]
> Go port was almost done by GitHub Copilot. Use with care.

> [!WARNING]
> Latest version is using [SIMD](https://github.com/pehringer/simd). If you need Pure Go implementation, use `v0.0.4-no-simd` or `no-simd`

## binary file format (gv)

```text
0: uint32_t width
4: uint32_t height
8: uint32_t frame count
12: float fps
16: uint32_t format (DXT1 = 1, DXT3 = 3, DXT5 = 5, BC7 = 7)
20: uint32_t frame bytes
24: raw frame storage (lz4 compressed)
eof - (frame count) * 16: [(uint64_t, uint64_t)..<frame count] (address, size) of lz4, address is zero based from file head
```

More detail, see [ofxExtremeGpuVideo](https://github.com/Ushio/ofxExtremeGpuVideo).

## Test

```bash
$ go test ./gvvideo
```
