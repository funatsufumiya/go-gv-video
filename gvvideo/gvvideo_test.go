package gvvideo

import (
	"path/filepath"
	"testing"
)

func almostEqual(a, b byte) bool {
	d := int(a) - int(b)
	if d < 0 {
		d = -d
	}
	return d <= 8
}

func assertRGBA(t *testing.T, got []uint8, wantR, wantG, wantB, wantA byte, msg string) {
	if !almostEqual(got[0], wantR) || !almostEqual(got[1], wantG) ||
		!almostEqual(got[2], wantB) || !almostEqual(got[3], wantA) {
		t.Errorf("%s: got %v, want [%d %d %d %d]", msg, got[:4], wantR, wantG, wantB, wantA)
	}
}

func TestGVVideo_ReadFrame(t *testing.T) {
	gvPath := filepath.Join("test_asset", "test-10px.gv")
	video, err := LoadGVVideo(gvPath)
	if err != nil {
		t.Fatalf("failed to load gv: %v", err)
	}
	w, h := int(video.Header.Width), int(video.Header.Height)
	if w != 10 || h != 10 {
		t.Errorf("unexpected size: %dx%d", w, h)
	}
	if video.Header.FrameCount != 5 {
		t.Errorf("unexpected frame count: %d", video.Header.FrameCount)
	}
	if video.Header.FPS != 1.0 {
		t.Errorf("unexpected fps: %f", video.Header.FPS)
	}
	if video.Header.Format != GVFormatDXT1 {
		t.Errorf("unexpected format: %d", video.Header.Format)
	}
	if video.Header.FrameBytes != 72 {
		t.Errorf("unexpected frame bytes: %d", video.Header.FrameBytes)
	}

	frame, err := video.ReadFrame(3)
	if err != nil {
		t.Fatalf("failed to read frame: %v", err)
	}
	if len(frame) != w*h*4 {
		t.Errorf("unexpected frame length: %d", len(frame))
	}
	assertRGBA(t, frame[:4], 255, 0, 0, 255, "(0,0) should be red")
	assertRGBA(t, frame[6*4:6*4+4], 0, 0, 255, 255, "(6,0) should be blue")
	assertRGBA(t, frame[(0+w*6)*4:(0+w*6)*4+4], 0, 255, 0, 255, "(0,6) should be green")
	assertRGBA(t, frame[(6+w*6)*4:(6+w*6)*4+4], 231, 255, 0, 255, "(6,6) should be yellow (allow error)")

	_, err = video.ReadFrame(5)
	if err == nil {
		t.Errorf("expected error for out-of-range frame")
	}
}

func TestGVVideo_ReadFrameAt(t *testing.T) {
	gvPath := filepath.Join("test_asset", "test-10px.gv")
	video, err := LoadGVVideo(gvPath)
	if err != nil {
		t.Fatalf("failed to load gv: %v", err)
	}
	frame, err := video.ReadFrame(0)
	if err != nil {
		t.Fatalf("failed to read frame at 0: %v", err)
	}
	assertRGBA(t, frame[:4], 255, 0, 0, 255, "(0,0) should be red")
}
