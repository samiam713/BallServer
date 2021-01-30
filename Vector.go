package main

import "math"

// Vector is a 2d float vector
type Vector struct {
	X, Y float64
}

func (v *Vector) subtracting(o Vector) {
	v.X -= o.X
	v.Y -= o.Y
}

func subtract(v1 Vector, v2 Vector) Vector {
	v1.subtracting(v2)
	return v1
}

func (v *Vector) adding(o Vector) {
	v.X += o.X
	v.Y += o.Y
}

func add(v1 Vector, v2 Vector) Vector {
	v1.adding(v2)
	return v1
}

func dot(v1, v2 *Vector) float64 {
	return v1.X*v2.X + v1.Y*v2.Y
}

func (v Vector) magnitude() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v *Vector) scale(s float64) {
	v.X *= s
	v.Y *= s
}

func (v Vector) scaledBy(s float64) Vector {
	v.scale(s)
	return v
}

func (v Vector) unit() Vector {
	if v.X == 0 && v.Y == 0 {
		return v
	}
	mag := v.magnitude()

	v.scale(1 / mag)
	return v
}

func (v *Vector) decomposed(direction *Vector) (Vector, Vector) {
	directionUnit := direction.unit()
	inDirection := dot(&directionUnit, v)
	directionUnit.scale(inDirection)
	return directionUnit, subtract(*v, directionUnit)
}
