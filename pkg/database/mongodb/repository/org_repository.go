package repository

import (
	"context"
	"errors"
	"fmt"

	"Go-api/pkg/database/mongodb/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrganizationRepository struct {
	db *mongo.Database
}

func NewOrganizationRepository(db *mongo.Database) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

func (r *OrganizationRepository) CreateOrganization(org *models.Organization) (string, error) {
	collection := r.db.Collection("organization")

	// Check if organization name already exists
	existingOrg := r.GetOrganizationByName(org.Name)
	if existingOrg != nil {
		return "", errors.New("organization name already exists")
	}

	res, err := collection.InsertOne(context.Background(), org)
	if err != nil {
		return "", fmt.Errorf("failed to create organization: %w", err)
	}
	// Retrieve the ID of the newly created organization
	organizationID := res.InsertedID.(primitive.ObjectID).Hex()
	return organizationID, nil
}

func (r *OrganizationRepository) GetOrganizationByName(name string) *models.Organization {
	var organization models.Organization
	collection := r.db.Collection("organization")
	err := collection.FindOne(context.Background(), bson.M{"name": name}).Decode(&organization)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil // Organization not found
		}
		// Handle other errors if needed
		return nil
	}
	return &organization
}

func (r *OrganizationRepository) UpdateOrganization(id string, org *models.Organization) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid organization ID")
	}

	collection := r.db.Collection("organization")
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": org}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}
	return nil
}

func (r *OrganizationRepository) DeleteOrganization(id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid organization ID")
	}

	collection := r.db.Collection("organization")
	filter := bson.M{"_id": objID}

	_, err = collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}
	return nil
}

func (r *OrganizationRepository) GetAllOrganizations() ([]models.Organization, error) {
	var organizations []models.Organization

	collection := r.db.Collection("organization")
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve organizations: %w", err)
	}
	defer cursor.Close(context.Background())

	err = cursor.All(context.Background(), &organizations)
	if err != nil {
		return nil, fmt.Errorf("failed to decode organizations: %w", err)
	}

	return organizations, nil
}

func (r *OrganizationRepository) GetOrganizationByID(id string) (*models.Organization, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid organization ID")
	}

	var organization models.Organization

	collection := r.db.Collection("organization")
	err = collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&organization)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Organization not found
		}
		return nil, fmt.Errorf("failed to retrieve organization: %w", err)
	}

	return &organization, nil
}

func (r *OrganizationRepository) GetAccessLevelByEmail(organizationID, email string) (int, error) {
	objID, err := primitive.ObjectIDFromHex(organizationID)
	if err != nil {
		return -1, errors.New("invalid organization ID")
	}

	collection := r.db.Collection("organization")
	filter := bson.M{
		"_id":                        objID,
		"organization_members.email": email,
	}

	var result struct {
		OrganizationMembers []models.OrganizationMember `bson:"organization_members"`
	}

	err = collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return -1, nil // Email not found, return access level -1
		}
		return -1, fmt.Errorf("failed to retrieve access level: %w", err)
	}

	if len(result.OrganizationMembers) == 0 {
		return -1, nil // Email not found, return access level -1
	}

	return result.OrganizationMembers[0].AccessLevel, nil
}

func (r *OrganizationRepository) AddMember(organizationID string, member *models.OrganizationMember) error {
	objID, err := primitive.ObjectIDFromHex(organizationID)
	if err != nil {
		return errors.New("invalid organization ID")
	}

	collection := r.db.Collection("organization")
	filter := bson.M{"_id": objID, "organization_members.email": bson.M{"$ne": member.Email}}

	// Check if the email already exists in the organization
	var result struct {
		OrganizationMembers []models.OrganizationMember `bson:"organization_members"`
	}

	err = collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("email already exists in the organization")
		}
		return fmt.Errorf("failed to check existing member: %w", err)
	}

	// If the email doesn't exist, proceed to add the member
	update := bson.M{"$push": bson.M{"organization_members": member}}

	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": objID}, update)
	if err != nil {
		return fmt.Errorf("failed to add member to organization: %w", err)
	}
	return nil
}
