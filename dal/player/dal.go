package player

import (
	"fmt"

	"github.com/byyjoww/leaderboard/constants"
	"github.com/go-pg/pg"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type PlayerDAL interface {
	Exists(name string) (bool, error)
	GetByPK(id uuid.UUID) (*Player, error)
	GetRankedByPK(id uuid.UUID) (*RankedPlayer, error)
	List(leaderboardId uuid.UUID, limit int, offset int) ([]*RankedPlayer, error)
	Create(player *Player) error
	UpdateScore(player *Player) error
	Delete(player *Player) error
}

type DAL struct {
	db *pg.DB
}

func NewDAL(db *pg.DB) *DAL {
	return &DAL{
		db: db,
	}
}

func (d *DAL) Exists(name string) (bool, error) {
	player := &Player{Name: name}
	return d.db.Model(player).
		Where("name = ?", name).
		Exists()
}

func (d *DAL) GetByPK(id uuid.UUID) (*Player, error) {
	player := &Player{ID: id}
	err := d.db.Model(player).
		WherePK().
		Select()
	if err != nil {
		return nil, err
	}
	return player, nil
}

func (d *DAL) GetRankedByPK(id uuid.UUID) (*RankedPlayer, error) {
	player := &RankedPlayer{Player: &Player{ID: id}}
	err := d.db.Model(player).
		WherePK().
		Column("*").
		ColumnExpr("Row_number() OVER (PARTITION BY leaderboard_id ORDER BY score DESC) as rank").
		Select()
	if err != nil {
		return nil, err
	}
	return player, nil
}

func (d *DAL) List(leaderboardId uuid.UUID, limit int, offset int) ([]*RankedPlayer, error) {
	var players []*RankedPlayer
	err := d.db.Model(&players).
		Where("leaderboard_id = ?", leaderboardId).
		Order("score DESC").
		Column("*").
		ColumnExpr("Row_number() OVER (PARTITION BY leaderboard_id ORDER BY score DESC) as rank").
		Limit(limit).
		Offset(offset).
		Select()
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}
	return players, nil
}

func (d *DAL) Create(player *Player) error {
	_, err := d.db.Model(player).
		Set("created_at = now()").
		Set("updated_at = now()").
		Insert()
	if err != nil {
		return err
	}
	return nil
}

func (d *DAL) UpdateScore(player *Player) error {
	_, err := d.db.Model(player).
		WherePK().
		Set("score = ?", player.Score).
		Set("updated_at = now()").
		Update()
	if err != nil {
		if err == pg.ErrNoRows {
			return errors.Wrap(
				constants.ErrPlayerNotFound,
				fmt.Sprintf("%s (id %s)", err.Error(), player.ID),
			)
		}
		return err
	}
	return nil
}

func (d *DAL) Delete(player *Player) error {
	_, err := d.db.Model(player).
		WherePK().
		Delete()
	if err != nil {
		if err == pg.ErrNoRows {
			return errors.Wrap(constants.ErrPlayerNotFound, err.Error())
		}
		return err
	}
	return nil
}
