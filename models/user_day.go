package models

import "time"

type UserDay struct {
	Id         int       `xorm:"INT pk autoincr 'id'"`
	Uid        int       `xorm:"INT 'uid'"`
	DAY        time.Time `XORM:"DATETIME 'day'"`
	Num        int       `xorm:"INT 'num'"`
	SysCreated time.Time `xorm:"DATETIME DEFAULT CURRENT_TIMESTAMP created 'sys_created'"`
	SysUpdated time.Time `xorm:"DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP updated 'sys_updated'"`
}

func (this *UserDay) TableName() string {
	return "user_day"
}
