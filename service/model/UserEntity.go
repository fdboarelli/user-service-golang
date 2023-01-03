package model

// User is the user model, with bson and json identifiers for marshaling
type User struct {
	ID        string `bson:"id" json:"id"`
	Firstname string `bson:"first_name" json:"first_name"`
	Lastname  string `bson:"last_name" json:"last_name"`
	Nickname  string `bson:"nickname" json:"nickname"`
	Password  string `bson:"password" json:"password"`
	Email     string `bson:"email" json:"email"`
	Country   string `bson:"country" json:"country"`
	CreatedAt string `bson:"created_at" json:"created_at"`
	UpdatedAt string `bson:"updated_at" json:"updated_at"`
}
