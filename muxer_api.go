package gohlslib

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
)

func (s *muxerServer) LastIDs() (*uint64, *uint64) {
	var segID, partID *uint64
	if len(s.segments) > 0 {
		seg := s.segments[len(s.segments)-1]
		id, _ := extractID(seg.getName())
		segID = &id
	}
	if len(s.nextSegmentParts) > 0 {
		part := s.nextSegmentParts[len(s.nextSegmentParts)-1]
		id, _ := extractID(part.getName())
		partID = &id
	}

	return segID, partID
}

func (s *muxerServer) Index() []byte {
	return s.multivariantPlaylist
}

func (s *muxerServer) InitFile() ([]byte, error) {
	f, err := generateInitFile(s.videoTrack, s.audioTrack, s.storageFactory, s.prefix)
	if err != nil {
		return nil, err
	} else {
		var b []byte
		rc, err := f.Reader()
		if err != nil {
			return nil, err
		} else {
			defer rc.Close()

			b, err = io.ReadAll(rc)
			if err != nil {
				return nil, err
			}

			return b, nil
		}
	}
}

func (s *muxerServer) Prefix() string {
	return s.prefix
}

func (s *muxerServer) Playlist(isDeltaUpdate bool) ([]byte, error) {
	return generateMediaPlaylist(
		isDeltaUpdate,
		s.variant,
		s.segments,
		s.nextSegmentParts,
		s.nextPartID,
		s.segmentDeleteCount,
		s.prefix,
	)
}

func (s *muxerServer) Segment(id uint64) ([]byte, error) {
	fname := fmt.Sprintf("%s_seg%d.mp4", s.prefix, id)
	s.mutex.Lock()
	segment, ok := s.segmentsByName[fname]
	s.mutex.Unlock()
	if !ok {
		return nil, errors.New("not found")
	}

	r, err := segment.reader()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (s *muxerServer) Part(id uint64) ([]byte, error) {
	fname := fmt.Sprintf("%s_part%d.mp4", s.prefix, id)
	s.mutex.Lock()
	part, ok := s.partsByName[fname]
	s.mutex.Unlock()
	if !ok {
		return nil, errors.New("not found")
	}

	r, err := part.reader()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return b, nil
}

var re = regexp.MustCompile(`_(seg|part)(\d+)\.mp4$`)

func extractID(name string) (uint64, error) {
	matches := re.FindStringSubmatch(name)

	if len(matches) < 3 {
		return 0, fmt.Errorf("no ID found in '%s'", name)
	}

	id, err := strconv.Atoi(matches[2])
	if err != nil {
		return 0, fmt.Errorf("error parsing ID from '%s': %v", name, err)
	}

	return uint64(id), nil
}
