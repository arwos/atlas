package database

import (
	"go.osspkg.com/goppy/v3/plugins/orm"
)

const (
	tagMaster = "master"
	tagSlave  = "slave"
)

type Service struct {
	db orm.ORM
}

func NewService(db orm.ORM) *Service {
	return &Service{
		db: db,
	}
}

func (s *Service) Master() orm.Stmt {
	return s.db.Tag(tagMaster)
}

func (s *Service) Slave() orm.Stmt {
	return s.db.Tag(tagSlave)
}
