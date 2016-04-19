package conjgrad

import (
	"math"

	"github.com/unixpickle/num-analysis/linalg"
)

const residualUpdateFrequency = 20

// SolveStoppable solves a system of linear
// equations t*x = b for x, where t is a
// symmetric positive-definite operator.
//
// The prec argument specifies a bound on the
// residual error of the solution. If the largest
// element of (Ax-b) has an absolute value less than
// prec, then the current x is returned.
//
// The cancelChan argument is a channel which you
// can close to stop the solve early.
// If the solve is cancelled, an approximate
// solution is returned.
func SolveStoppable(t LinTran, b linalg.Vector, prec float64,
	cancelChan <-chan struct{}) linalg.Vector {
	var conjVec linalg.Vector
	var residual linalg.Vector
	var solution linalg.Vector

	var lastResidualDot float64

	residual = b.Copy()
	solution = make(linalg.Vector, t.Dim())

	for i := 0; i < t.Dim(); i++ {
		if greatestValue(residual) <= prec {
			break
		}
		if i == 0 {
			conjVec = residual.Copy()
			lastResidualDot = conjVec.Dot(conjVec)
		} else {
			residualDot := residual.Dot(residual)
			projAmount := -residualDot / lastResidualDot
			lastResidualDot = residualDot
			conjVec = residual.Copy().Add(conjVec.Scale(-projAmount))
		}
		if allZero(conjVec) {
			break
		}
		optimalDistance := conjVec.Dot(residual) / conjVec.Dot(t.Apply(conjVec))

		solution.Add(conjVec.Copy().Scale(optimalDistance))
		if i != 0 && (i%residualUpdateFrequency) == 0 {
			residual = t.Apply(solution).Scale(-1).Add(b)
		} else {
			residual.Add(t.Apply(conjVec).Scale(-optimalDistance))
		}

		select {
		case <-cancelChan:
			return solution
		default:
		}
	}

	return solution
}

// SolvePrec is like SolveStoppable, but it does not
// give you the option to cancel the solve early.
func SolvePrec(t LinTran, b linalg.Vector, prec float64) linalg.Vector {
	return SolveStoppable(t, b, prec, nil)
}

// Solve is like SolvePrec, except that it computes
// as accurate a solution as possible.
func Solve(t LinTran, b linalg.Vector) linalg.Vector {
	return SolvePrec(t, b, 0)
}

func allZero(v linalg.Vector) bool {
	for _, x := range v {
		if x != 0 {
			return false
		}
	}
	return true
}

func greatestValue(v linalg.Vector) float64 {
	var max float64
	for _, x := range v {
		max = math.Max(max, math.Abs(x))
	}
	return max
}