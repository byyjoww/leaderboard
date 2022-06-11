package leaderboard

import (
	"github.com/byyjoww/leaderboard/constants"
	"github.com/byyjoww/leaderboard/dal/player"
	"github.com/go-pg/pg"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type LeaderboardDAL interface {
	GetByPK(id uuid.UUID) (*Leaderboard, error)
	List() ([]*Leaderboard, error)
	Create(leaderboard *Leaderboard) error
	Delete(leaderboard *Leaderboard) error
}

type DAL struct {
	db *pg.DB
}

func NewDAL(db *pg.DB) *DAL {
	return &DAL{
		db: db,
	}
}

func (d *DAL) GetByPK(id uuid.UUID) (*Leaderboard, error) {
	leaderboard := &Leaderboard{ID: id}
	err := d.db.Model(leaderboard).
		WherePK().
		Select()
	if err != nil {
		return nil, err
	}
	return leaderboard, nil
}

func (d *DAL) List() ([]*Leaderboard, error) {
	var leaderboards []*Leaderboard
	err := d.db.Model(&leaderboards).
		Select()
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}
	return leaderboards, nil
}

func (d *DAL) Create(leaderboard *Leaderboard) error {
	_, err := d.db.Model(leaderboard).
		Set("created_at = now()").
		Set("updated_at = now()").
		Insert()
	if err != nil {
		return err
	}
	return nil
}

func (d *DAL) Delete(leaderboard *Leaderboard) error {
	return d.db.RunInTransaction(func(tx *pg.Tx) error {
		// List all players in the leaderboard
		var players []*player.Player
		err := d.db.Model(&players).
			Where("leaderboard_id = ?", leaderboard.ID).
			Select()
		if err != nil && err != pg.ErrNoRows {
			return err
		}

		// Delete all players in the leaderboard
		err = d.db.Delete(&players)
		if err != nil {
			if err == pg.ErrNoRows {
				return errors.Wrap(constants.ErrPlayerNotFound, err.Error())
			}
			return err
		}

		// Delete the leaderboard
		_, err = d.db.Model(leaderboard).
			WherePK().
			Delete()
		if err != nil {
			if err == pg.ErrNoRows {
				return errors.Wrap(constants.ErrLeaderboardNotFound, err.Error())
			}
			return err
		}

		return nil
	})
}