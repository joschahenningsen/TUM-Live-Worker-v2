package model

import "time"

type StreamRequest struct {
	SourceType       string    `json:"source_type"`
	SourceUrl        string    `json:"source_url"`
	CourseSlug       string    `json:"course_slug"`
	Year             uint      `json:"year"`
	TeachingTerm     string    `json:"teaching_term"`
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	PublishStream    bool      `json:"publish_stream"`
	PublishRecording bool      `json:"publish_recording"`
}
