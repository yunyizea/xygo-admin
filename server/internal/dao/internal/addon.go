// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AddonDao is the data access object for the table xy_addon.
type AddonDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  AddonColumns       // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// AddonColumns defines and stores column names for the table xy_addon.
type AddonColumns struct {
	Id            string //
	Name          string //
	Version       string //
	Title         string //
	Status        string //
	InstalledAt   string //
	UninstalledAt string //
}

// addonColumns holds the columns for the table xy_addon.
var addonColumns = AddonColumns{
	Id:            "id",
	Name:          "name",
	Version:       "version",
	Title:         "title",
	Status:        "status",
	InstalledAt:   "installed_at",
	UninstalledAt: "uninstalled_at",
}

// NewAddonDao creates and returns a new DAO object for table data access.
func NewAddonDao(handlers ...gdb.ModelHandler) *AddonDao {
	return &AddonDao{
		group:    "default",
		table:    "xy_addon",
		columns:  addonColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AddonDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AddonDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AddonDao) Columns() AddonColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AddonDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AddonDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *AddonDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
