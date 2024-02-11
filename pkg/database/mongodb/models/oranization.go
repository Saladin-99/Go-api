package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Organization represents an organization
type Organization struct {
	ID                  primitive.ObjectID   `bson:"_id,omitempty"`
	Name                string               `bson:"name"`
	Description         string               `bson:"description"`
	OrganizationMembers []OrganizationMember `bson:"organization_members,omitempty"`
}

type OrganizationMember struct {
	Name        string `bson:"name"`
	Email       string `bson:"email"`
	AccessLevel int    `bson:"access_level"`
}
