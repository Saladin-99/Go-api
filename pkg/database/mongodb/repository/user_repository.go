package repository

import (
	"context"
	"errors"

	"Go-api/pkg/database/mongodb/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	db *mongo.Database
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(user *models.User) error {
	// Hash the password before storing it in the database
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	_, err = r.db.Collection("user").InsertOne(context.Background(), user)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) UpdateUser(id string, user *models.User) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid user ID")
	}

	// Hash the password before updating it in the database, if provided
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(hashedPassword)
	}

	update := bson.M{
		"$set": bson.M{
			"name":     user.Name,
			"email":    user.Email,
			"password": user.Password,
		},
	}

	_, err = r.db.Collection("user").UpdateOne(context.Background(), bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) DeleteUser(id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid user ID")
	}

	_, err = r.db.Collection("user").DeleteOne(context.Background(), bson.M{"_id": objID})
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) GetUser(id string) (*models.User, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	var user models.User
	err = r.db.Collection("user").FindOne(context.Background(), bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // User not found
		}
		return nil, err
	}

	return &user, nil
}
func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Collection("user").FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // User not found
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) AuthenticateUser(email, password string) (*models.User, error) {
	var user models.User
	err := r.db.Collection("user").FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Compare the provided password with the hashed password from the database
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid password")
	}

	return &user, nil
}
