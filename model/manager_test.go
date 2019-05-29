package model

import (
	"fmt"
	"testing"
	"time"
)

func TestManager_Broad(t *testing.T) {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <- ticker.C:
			fmt.Println(1)
		}
	}
}