package main

import (
	"github.com/urfave/cli/v2"
	"os"
	"time"

	"github.com/looplab/fsm"
	"github.com/rs/zerolog/log"
)

type PomodoroClock struct {
	Cycles          int // Cycles to run
	CompletedCycles int
	WorkTime        time.Duration // Minutes to work
	ShortRestTime   time.Duration // Minutes to short rest
	LongRestTime    time.Duration // Minutes to long rest
	State           *fsm.FSM
}

func (pm *PomodoroClock) working(e *fsm.Event) {
	if pm.Cycles <= 0 {
		e.FSM.SetState("end")
		return
	}

	log.Debug().Int("CompletedCycles", pm.CompletedCycles).Str("state", e.FSM.Current()).Dur("timeout", pm.WorkTime).Msg("Working for...")
	waitForTime(pm.WorkTime)
	pm.State.SetState("rest")

}

func (pm *PomodoroClock) resting(e *fsm.Event) {
	var timeout time.Duration
	if (pm.CompletedCycles+1)%4 == 0 {
		timeout = pm.LongRestTime
	} else {
		timeout = pm.ShortRestTime
	}

	log.Debug().Int("CompletedCycles", pm.CompletedCycles).Str("state", e.FSM.Current()).Dur("timeout", timeout).Msg("Sleeping for...")
	waitForTime(timeout)
	pm.State.SetState("begin")
	pm.CompletedCycles += 1
}

func (pm *PomodoroClock) finalizing(e *fsm.Event) {
	log.Debug().Msg("finalizing callback")
}

func (pm *PomodoroClock) RunPomodoro(c *cli.Context) error {

	for pm.CompletedCycles < pm.Cycles {

		switch pm.State.Current() {
		case "begin":
			err := pm.State.Event("working")
			if err != nil {
				return err
			}

		case "rest":
			err := pm.State.Event("resting")
			if err != nil {
				return err
			}

		case "end":
			err := pm.State.Event("finalizing")
			if err != nil {
				return err
			}
		}
	}
	return nil

}

func NewPomodoroClock() *PomodoroClock {

	pm := &PomodoroClock{}
	pm.State = fsm.NewFSM(
		"begin",
		fsm.Events{
			{Name: "working", Src: []string{"begin"}, Dst: "work"},
			{Name: "resting", Src: []string{"rest"}, Dst: "work"},
			{Name: "finalizing", Src: []string{"end"}, Dst: "begin"},
		},
		fsm.Callbacks{
			"working":    pm.working,
			"resting":    pm.resting,
			"finalizing": pm.finalizing,
		},
	)

	return pm
}

func waitForTime(timeout time.Duration) {
	select {
	case <-time.After(timeout * time.Minute):
	}
}
func main() {

	pm := NewPomodoroClock()

	app := &cli.App{
		Name:  "run",
		Usage: "run pomodoro clock",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:        "CompletedCycles",
				Value:       0,
				Usage:       "number of completed cycles",
				Destination: &pm.CompletedCycles,
			},
			&cli.IntFlag{
				Name:        "Cycles",
				Value:       1,
				Usage:       "number of cycles",
				Destination: &pm.Cycles,
			},
			&cli.DurationFlag{
				Name:        "ShortRestTime",
				Value:       3,
				Usage:       "short time to rest",
				Destination: &pm.ShortRestTime,
			},
			&cli.DurationFlag{
				Name:        "LongRestTime",
				Value:       15,
				Usage:       "long time to rest",
				Destination: &pm.LongRestTime,
			},
			&cli.DurationFlag{
				Name:        "WorkTime",
				Value:       25,
				Usage:       "time to work",
				Destination: &pm.WorkTime,
			},
		},
		Action: pm.RunPomodoro,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal().AnErr("err", err)

	}
}
