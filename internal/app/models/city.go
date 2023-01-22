package models

type CityID int

type City struct {
	ID   CityID `json:"id"`
	Name string `json:"name"`
}
