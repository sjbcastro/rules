package rules

import (
	"errors"
)

type TeamRuleset struct {
	StandardRuleset

	TeamMap map[string]string

	// These are intentionally designed so that they default to a standard game.
	AllowBodyCollisions bool
	SharedElimination   bool
	SharedHealth        bool
	SharedLength        bool
}

const EliminatedByTeam = "team-eliminated"

func (r *TeamRuleset) CreateNextBoardState(prevState *BoardState, moves []SnakeMove) (*BoardState, error) {
	nextBoardState, err := r.StandardRuleset.CreateNextBoardState(prevState, moves)
	if err != nil {
		return nil, err
	}

	// TODO: LOG?
	err = r.resurrectTeamBodyCollisions(nextBoardState)
	if err != nil {
		return nil, err
	}

	// TODO: LOG?
	err = r.shareTeamAttributes(nextBoardState)
	if err != nil {
		return nil, err
	}

	return nextBoardState, nil
}

func (r *TeamRuleset) areSnakesOnSameTeam(snake *Snake, other *Snake) bool {
	return r.areSnakeIDsOnSameTeam(snake.ID, other.ID)
}

func (r *TeamRuleset) areSnakeIDsOnSameTeam(snakeID string, otherID string) bool {
	return r.TeamMap[snakeID] == r.TeamMap[otherID]
}

func (r *TeamRuleset) resurrectTeamBodyCollisions(b *BoardState) error {
	if !r.AllowBodyCollisions {
		return nil
	}

	for i := 0; i < len(b.Snakes); i++ {
		snake := &b.Snakes[i]
		if snake.EliminatedCause == EliminatedByCollision {
			if snake.EliminatedBy == "" {
				return errors.New("snake eliminated by collision and eliminatedby is not set")
			}
			if snake.ID != snake.EliminatedBy && r.areSnakeIDsOnSameTeam(snake.ID, snake.EliminatedBy) {
				snake.EliminatedCause = NotEliminated
				snake.EliminatedBy = ""
			}
		}
	}

	return nil
}

func (r *TeamRuleset) shareTeamAttributes(b *BoardState) error {
	if !(r.SharedElimination || r.SharedLength || r.SharedHealth) {
		return nil
	}

	for i := 0; i < len(b.Snakes); i++ {
		snake := &b.Snakes[i]
		if snake.EliminatedCause != NotEliminated {
			continue
		}

		for j := 0; j < len(b.Snakes); j++ {
			other := &b.Snakes[j]
			if r.areSnakesOnSameTeam(snake, other) {
				if r.SharedHealth {
					if snake.Health < other.Health {
						snake.Health = other.Health
					}
				}
				if r.SharedLength {
					if len(snake.Body) == 0 || len(other.Body) == 0 {
						return errors.New("found snake of zero length")
					}
					for len(snake.Body) < len(other.Body) {
						r.growSnake(snake)
					}
				}
				if r.SharedElimination {
					if snake.EliminatedCause == NotEliminated && other.EliminatedCause != NotEliminated {
						snake.EliminatedCause = EliminatedByTeam
						// We intentionally do not set snake.EliminatedBy because there might be multiple culprits.
						snake.EliminatedBy = ""
					}
				}
			}
		}
	}

	return nil
}

func (r *TeamRuleset) IsGameOver(b *BoardState) (bool, error) {
	snakesRemaining := []*Snake{}
	for i := 0; i < len(b.Snakes); i++ {
		if b.Snakes[i].EliminatedCause == NotEliminated {
			snakesRemaining = append(snakesRemaining, &b.Snakes[i])
		}
	}

	for i := 0; i < len(snakesRemaining); i++ {
		if !r.areSnakesOnSameTeam(snakesRemaining[i], snakesRemaining[0]) {
			// There are multiple teams remaining
			return false, nil
		}
	}
	// no snakes or single team remaining
	return true, nil
}
