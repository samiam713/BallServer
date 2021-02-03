package main

const (
	ballRadius   = 0.05
	ballDiameter = ballRadius * 2
)

// Ball represents all data necessary for a ball
type Ball struct {
	pos, vel, accAccum Vector
}

var hostBall = Ball{pos: Vector{0, ballInitialY}, vel: Vector{0, 0}}
var clientBall = Ball{pos: Vector{0, -ballInitialY}, vel: Vector{0, 0}}

func (ball *Ball) advanceBall(dt float64) {
	ball.vel.adding(ball.accAccum)
	// (0.995)^60 is approximately 0.75
	ball.vel.scale(0.995)
	ball.accAccum.X = 0
	ball.accAccum.Y = 0
	ball.pos.adding(ball.vel.scaledBy(dt))

}

func areColliding(b1, b2 *Ball) bool {
	return subtract(b1.pos, b2.pos).magnitude() < ballDiameter
}

func collideIfNecessary(b1, b2 *Ball) {
	if areColliding(b1, b2) {
		// find each's velocity component in direction of other
		b1Tob2 := subtract(b2.pos, b1.pos)
		b2Tob1 := b1Tob2.scaledBy(-1)

		b1Par, b1Perp := b1.vel.decomposed(&b1Tob2)
		b2Par, b2Perp := b2.vel.decomposed(&b2Tob1)

		// swap velocity components in direction of each other
		b1.vel = add(b1Perp, b2Par)
		b2.vel = add(b2Perp, b1Par)

		// make them barely kiss
		average := add(b1.pos, b2.pos).scaledBy(0.5)
		b1.pos = add(average, b2Tob1.unit().scaledBy(1.01*ballRadius))
		b2.pos = add(average, b1Tob2.unit().scaledBy(1.01*ballRadius))
	}
}
