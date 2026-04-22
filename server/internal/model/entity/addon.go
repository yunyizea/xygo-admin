// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Addon is the golang structure for table addon.
type Addon struct {
	Id            uint64 `json:"id"            orm:"id"             description:""` //
	Name          string `json:"name"          orm:"name"           description:""` //
	Version       string `json:"version"       orm:"version"        description:""` //
	Title         string `json:"title"         orm:"title"          description:""` //
	Status        int    `json:"status"        orm:"status"         description:""` //
	InstalledAt   uint64 `json:"installedAt"   orm:"installed_at"   description:""` //
	UninstalledAt uint64 `json:"uninstalledAt" orm:"uninstalled_at" description:""` //
}
