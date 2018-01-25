package main

import (
	"errors"
	"testing"
)

func TestLoadContext(t *testing.T) {
	t.Run("NoErrors", func(t *testing.T) {
		lc := NewErrorContext()

		lc.Add(2)
		lc.Done()
		lc.Done()

		err := lc.Wait()

		if err != nil {
			t.Fatalf("Error is not nil, is %v", err)
		}

	})

	t.Run("OneError", func(t *testing.T) {
		lc := NewErrorContext()

		lc.Add(256)
		lc.Cancel(errors.New("Some error or something"))

		err := lc.Wait()

		if err == nil {
			t.Fatal("Error is nil")
		}

	})

	t.Run("SeveralErrors", func(t *testing.T) {
		lc := NewErrorContext()

		err1 := errors.New("The first err")
		err2 := errors.New("The second err")

		lc.Add(256)
		lc.Cancel(err1)
		lc.Cancel(err2)

		err := lc.Wait()

		if err != err1 {
			t.Fatalf("Err was not %v, was %v", err1, err)
		}
	})

	t.Run("CancelingWork", func(t *testing.T) {
		lc := NewErrorContext()
		lc.Add(2)

		// Before it's canceled, lc.Canceled is open
		select {
		case <-lc.Canceled:
			t.Fatal("lc.Canceled is closed before lc is canceled.")
		default:
			// Everything's fine.
		}

		// Cancel lc with an error
		lc.Cancel(errors.New("An error"))

		// After it's canceled, lc.Canceled is closed
		select {
		case <-lc.Canceled:
			// Do nothing, this is fine.
		default:
			t.Fatal("lc.Canceled is not closed after lc is canceled.")
		}
	})
}
