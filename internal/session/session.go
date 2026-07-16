package session

import (
	"fmt"
	"time"

	"github.com/Ibnu-Afdel/pomogo/internal/timer"
)

// SegmentKind defines the type of focus segment.
type SegmentKind int

const (
	SegmentKindWork SegmentKind = iota
	SegmentKindShortBreak
	SegmentKindLongBreak
)

func (k SegmentKind) String() string {
	switch k {
	case SegmentKindWork:
		return "Work"
	case SegmentKindShortBreak:
		return "Short Break"
	case SegmentKindLongBreak:
		return "Long Break"
	default:
		return "Unknown"
	}
}

// Segment represents a discrete timing block.
type Segment struct {
	Kind     SegmentKind
	Duration time.Duration
}

// Mode represents the focus mode.
type Mode string

const (
	ModeQuick Mode = "quick"
	ModeDeep  Mode = "deep"
)

// Block coordinates segments for a focus session.
type Block struct {
	Mode           Mode
	Segments       []Segment
	Index          int
	PlannedTotal   time.Duration
	AutoAdvance    bool
	CurrentSegment Segment
	// Quick focus generator state
	SessionCount            int
	SessionsBeforeLongBreak int
	WorkDuration            time.Duration
	ShortBreakDuration      time.Duration
	LongBreakDuration       time.Duration
}

// BuildDeepPlan splits a total duration into Pomodoro-style segments.
func BuildDeepPlan(total, work, shortBreak, longBreak time.Duration, sessionsBeforeLongBreak int) []Segment {
	if total <= work {
		return []Segment{{Kind: SegmentKindWork, Duration: total}}
	}

	var segments []Segment
	remaining := total
	workSessions := 0

	for remaining > 0 {
		breakKind := SegmentKindShortBreak
		breakDuration := shortBreak
		if sessionsBeforeLongBreak > 0 && (workSessions+1)%sessionsBeforeLongBreak == 0 {
			breakKind = SegmentKindLongBreak
			breakDuration = longBreak
		}

		// Keep the block from ending on a break or a tiny trailing focus sliver.
		if remaining-work-breakDuration >= 10*time.Minute {
			segments = append(segments, Segment{Kind: SegmentKindWork, Duration: work})
			segments = append(segments, Segment{Kind: breakKind, Duration: breakDuration})
			workSessions++
			remaining -= work + breakDuration
		} else {
			// This is the final work segment, fold all remaining time here
			segments = append(segments, Segment{Kind: SegmentKindWork, Duration: remaining})
			remaining = 0
		}
	}

	return segments
}

// NewQuickBlock initializes a Quick Focus block.
func NewQuickBlock(work, short, long time.Duration, beforeLong int, autoAdvance bool) *Block {
	return &Block{
		Mode:                    ModeQuick,
		Index:                   0,
		AutoAdvance:             autoAdvance,
		SessionsBeforeLongBreak: beforeLong,
		WorkDuration:            work,
		ShortBreakDuration:      short,
		LongBreakDuration:       long,
		CurrentSegment:          Segment{Kind: SegmentKindWork, Duration: work},
	}
}

// NewDeepBlock initializes a Deep Focus block.
func NewDeepBlock(total, work, shortBreak, longBreak time.Duration, sessionsBeforeLongBreak int, autoAdvance bool) *Block {
	segs := BuildDeepPlan(total, work, shortBreak, longBreak, sessionsBeforeLongBreak)
	var current Segment
	if len(segs) > 0 {
		current = segs[0]
	}
	return &Block{
		Mode:                    ModeDeep,
		Segments:                segs,
		Index:                   0,
		PlannedTotal:            total,
		AutoAdvance:             autoAdvance,
		SessionsBeforeLongBreak: sessionsBeforeLongBreak,
		WorkDuration:            work,
		ShortBreakDuration:      shortBreak,
		LongBreakDuration:       longBreak,
		CurrentSegment:          current,
	}
}

// Remaining returns the overall remaining time of the block.
func (b *Block) Remaining(segRemaining time.Duration) time.Duration {
	if b.Mode == ModeQuick {
		return segRemaining
	}
	if b.Index < 0 || b.Index >= len(b.Segments) {
		return 0
	}
	rem := segRemaining
	for i := b.Index + 1; i < len(b.Segments); i++ {
		rem += b.Segments[i].Duration
	}
	return rem
}

// Advance moves the block to the next segment.
func (b *Block) Advance() (Segment, bool) {
	b.Index++
	if b.Mode == ModeDeep {
		if b.Index >= len(b.Segments) {
			return Segment{}, true
		}
		b.CurrentSegment = b.Segments[b.Index]
		return b.CurrentSegment, false
	}

	// ModeQuick cycle
	if b.CurrentSegment.Kind == SegmentKindWork {
		b.SessionCount++
		if b.SessionsBeforeLongBreak > 0 && b.SessionCount%b.SessionsBeforeLongBreak == 0 {
			b.CurrentSegment = Segment{Kind: SegmentKindLongBreak, Duration: b.LongBreakDuration}
		} else {
			b.CurrentSegment = Segment{Kind: SegmentKindShortBreak, Duration: b.ShortBreakDuration}
		}
	} else {
		b.CurrentSegment = Segment{Kind: SegmentKindWork, Duration: b.WorkDuration}
	}
	return b.CurrentSegment, false
}

// Progress computes the overall progress of the block.
func (b *Block) Progress(segRemaining time.Duration) float64 {
	if b.Mode == ModeQuick {
		// Quick mode uses segment-level progress
		return 0.0
	}
	if b.PlannedTotal <= 0 {
		return 0.0
	}
	rem := b.Remaining(segRemaining)
	elapsed := b.PlannedTotal - rem
	if elapsed < 0 {
		elapsed = 0
	}
	prog := float64(elapsed) / float64(b.PlannedTotal)
	if prog > 1.0 {
		prog = 1.0
	}
	return prog
}

// EventType defines the type of event returned from Runner.
type EventType int

const (
	EventNone EventType = iota
	EventSegmentEnded
	EventBlockEnded
)

// RunnerEvent details a timer runner event.
type RunnerEvent struct {
	Type    EventType
	Segment Segment
}

// Runner coordinates the block and active timer.Session.
type Runner struct {
	Block *Block
	Timer *timer.Session
}

// NewRunner creates a new Runner driving the block.
func NewRunner(block *Block) *Runner {
	t := timer.NewSession(0, 0, 0, 0)
	r := &Runner{
		Block: block,
		Timer: t,
	}
	r.ConfigureTimerForSegment(block.CurrentSegment, time.Time{})
	return r
}

// ConfigureTimerForSegment updates the underlying timer configuration.
func (r *Runner) ConfigureTimerForSegment(seg Segment, now time.Time) {
	r.Timer.RemainingTime = seg.Duration
	r.Timer.EndsAt = now.Add(seg.Duration)
	r.Timer.StartedAt = now

	switch seg.Kind {
	case SegmentKindWork:
		r.Timer.Phase = timer.PhaseWork
		if r.Timer.IsRunning {
			r.Timer.State = timer.StateWork
		} else {
			r.Timer.State = timer.StateIdle
		}
	case SegmentKindShortBreak:
		r.Timer.Phase = timer.PhaseShortBreak
		if r.Timer.IsRunning {
			r.Timer.State = timer.StateShortBreak
		} else {
			r.Timer.State = timer.StateIdle
		}
	case SegmentKindLongBreak:
		r.Timer.Phase = timer.PhaseLongBreak
		if r.Timer.IsRunning {
			r.Timer.State = timer.StateLongBreak
		} else {
			r.Timer.State = timer.StateIdle
		}
	}
}

// Start starts the current segment.
func (r *Runner) Start(clock timer.Clock) error {
	if r.Timer.IsRunning {
		return fmt.Errorf("runner already running")
	}
	now := clock.Now()
	r.Timer.IsRunning = true
	r.Timer.IsPaused = false
	r.Timer.StartedAt = now
	r.Timer.EndsAt = now.Add(r.Timer.RemainingTime)

	switch r.Timer.Phase {
	case timer.PhaseWork:
		r.Timer.State = timer.StateWork
	case timer.PhaseShortBreak:
		r.Timer.State = timer.StateShortBreak
	case timer.PhaseLongBreak:
		r.Timer.State = timer.StateLongBreak
	}

	return nil
}

// Tick checks progress and advances if segment ends.
func (r *Runner) Tick(clock timer.Clock) (RunnerEvent, bool) {
	if !r.Timer.IsRunning || r.Timer.IsPaused {
		return RunnerEvent{}, false
	}

	if r.Timer.Tick(clock) {
		endedSeg := r.Block.CurrentSegment
		nextSeg, blockEnded := r.Block.Advance()
		if blockEnded {
			r.Timer.IsRunning = false
			r.Timer.IsPaused = false
			r.Timer.State = timer.StateIdle
			return RunnerEvent{Type: EventBlockEnded, Segment: endedSeg}, true
		}

		r.ConfigureTimerForSegment(nextSeg, clock.Now())

		if r.Block.AutoAdvance {
			r.Timer.IsRunning = true
			r.Timer.IsPaused = false
			switch r.Timer.Phase {
			case timer.PhaseWork:
				r.Timer.State = timer.StateWork
			case timer.PhaseShortBreak:
				r.Timer.State = timer.StateShortBreak
			case timer.PhaseLongBreak:
				r.Timer.State = timer.StateLongBreak
			}
		} else {
			r.Timer.IsRunning = false
			r.Timer.IsPaused = false
			r.Timer.State = timer.StateIdle
		}

		return RunnerEvent{Type: EventSegmentEnded, Segment: endedSeg}, true
	}

	return RunnerEvent{}, false
}

// Skip advances the segment immediately.
func (r *Runner) Skip(clock timer.Clock) (RunnerEvent, bool) {
	endedSeg := r.Block.CurrentSegment
	nextSeg, blockEnded := r.Block.Advance()
	if blockEnded {
		r.Timer.IsRunning = false
		r.Timer.IsPaused = false
		r.Timer.State = timer.StateIdle
		return RunnerEvent{Type: EventBlockEnded, Segment: endedSeg}, true
	}

	r.ConfigureTimerForSegment(nextSeg, clock.Now())

	if r.Block.AutoAdvance {
		r.Timer.IsRunning = true
		r.Timer.IsPaused = false
		switch r.Timer.Phase {
		case timer.PhaseWork:
			r.Timer.State = timer.StateWork
		case timer.PhaseShortBreak:
			r.Timer.State = timer.StateShortBreak
		case timer.PhaseLongBreak:
			r.Timer.State = timer.StateLongBreak
		}
	} else {
		r.Timer.IsRunning = false
		r.Timer.IsPaused = false
		r.Timer.State = timer.StateIdle
	}
	return RunnerEvent{Type: EventSegmentEnded, Segment: endedSeg}, true
}
