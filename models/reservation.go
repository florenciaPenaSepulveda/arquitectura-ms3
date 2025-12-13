package models

import "time"

type Reservation struct {
  ID          uint      `gorm:"primaryKey" json:"id"`
  UserID      string    `gorm:"not null" json:"user_id"`
  CarID       string    `gorm:"not null" json:"car_id"`
  MechanicID  *string   `gorm:"" json:"mechanic_id"` // nullable
  Date        string    `gorm:"not null" json:"date"`
  Description string    `gorm:"type:text" json:"description"`
  Status      string    `gorm:"not null" json:"status"`
  CreatedAt   time.Time `json:"created_at"`
  UpdatedAt   time.Time `json:"updated_at"`
}

