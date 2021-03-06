package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/rakyll/portmidi"
	"github.com/simulatedsimian/joystick"
)

const (
	joystickIndex  = 0
	leftAxisIndex  = 3
	rightAxisIndex = 4
	stickAxisIndex = 5
)

func main() {
	js, err := joystick.Open(joystickIndex)
	if err != nil {
		log.Fatalln(err)
	}
	defer js.Close()

	portmidi.Initialize()
	defer portmidi.Terminate()

	deviceID := portmidi.DefaultOutputDeviceID()
	out, err := portmidi.NewOutputStream(deviceID, 1024, 0)
	if err != nil {
		js.Close()
		// TODO: return instead of Fatalln, so that defer is respected
		log.Fatalln(err)
	}
	defer out.Close()

	axisCount := js.AxisCount()

	if axisCount < 6 {
		js.Close()
		log.Fatalln("Joystick axis count < 6")
	}

	fmt.Printf("Joystick Name: %s\n", js.Name())
	fmt.Printf("Axis count: %d\n", axisCount)
	fmt.Printf("Button count: %d\n", js.ButtonCount())

	state, err := js.Read()
	if err != nil {
		js.Close()
		log.Fatalln(err)
	}

	var (
		minLeft         = 0
		maxLeft         = 0
		minRight        = 0
		maxRight        = 0
		minStick        = 0
		maxStick        = 0
		leftCalibrated  = false // once the range for the left light sensor is wide enough
		rightCalibrated = false // once the range for the right light sensor is wide enough
	)

	i := 0

	for {

		leftLight := state.AxisData[leftAxisIndex]
		rightLight := state.AxisData[rightAxisIndex]
		stick := state.AxisData[stickAxisIndex]

		if i == 0 {
			minLeft = leftLight
			maxLeft = leftLight
			minRight = rightLight
			maxRight = rightLight
			minStick = stick
			maxStick = stick
			i++
		} else {
			if leftLight < minLeft {
				minLeft = leftLight
			}
			if leftLight > maxLeft {
				maxLeft = leftLight
			}
			if rightLight < minRight {
				minRight = rightLight
			}
			if rightLight > maxRight {
				maxRight = rightLight
			}
			if stick < minStick {
				minStick = stick
			}
			if stick > maxStick {
				maxStick = stick
			}
		}

		leftP := 0.0
		leftRange := float64(maxLeft - minLeft)
		if leftRange > 0 {
			leftP = float64(leftLight-minLeft) / leftRange
		}

		rightP := 0.0
		rightRange := float64(maxRight - minRight)
		if rightRange > 0 {
			rightP = float64(rightLight-minRight) / rightRange
		}

		stickP := 0.0
		stickRange := float64(maxStick - minStick)
		if stickRange > 0 {
			stickP = float64(stick-minStick) / stickRange
		}

		//fmt.Printf("[l] min light max p [%d %d %d %v]\n", minLeft, leftLight, maxLeft, leftP)
		//fmt.Printf("[r] min light max [%d %d %d %v]\n", minRight, rightLight, maxRight, rightP)
		//fmt.Printf("[s] min stick max [%d %d %d %v]\n", minStick, stick, maxStick, stickP)

		// percentages, from 0 to 100, with lp and rp inverted
		lp := int(math.Round(100.0 + leftP*-100.0))
		rp := int(math.Round(100.0 + rightP*-100.0))
		sp := int(math.Round(stickP * 100))

		leftCalibrated = leftRange > 10000
		rightCalibrated = rightRange > 10000

		if lp > 25 && leftCalibrated {
			fmt.Println("Touched the left light sensor")
			fmt.Println("Outputting MIDI data")

			// note on events to play C major chord
			out.WriteShort(0x90, 60, 100)
			out.WriteShort(0x90, 64, 100)
			out.WriteShort(0x90, 67, 100)

			// notes will be sustained for 2 seconds
			time.Sleep(2 * time.Second)

			// note off events
			out.WriteShort(0x80, 60, 100)
			out.WriteShort(0x80, 64, 100)
			out.WriteShort(0x80, 67, 100)

		}
		if rp > 25 && rightCalibrated {
			fmt.Println("Touched the right light sensor")
			fmt.Println("Bye!")

			break
		}

		fmt.Printf("[l] %d\t[r] %d\t[s] %d\n", lp, rp, sp)

		time.Sleep(100 * time.Millisecond)
	}
}
