package controllers

import (
	"log"
	"net/http"

	"Go-api/pkg/database/mongodb/models"
	"Go-api/pkg/database/mongodb/repository"

	"github.com/gin-gonic/gin"
)

type OrganizationController struct {
	organizationRepository *repository.OrganizationRepository
	userRepository         *repository.UserRepository
	logger                 *log.Logger
}

func NewOrganizationController(logger *log.Logger, organizationRepository *repository.OrganizationRepository, userRepository *repository.UserRepository) *OrganizationController {
	return &OrganizationController{
		organizationRepository: organizationRepository,
		userRepository:         userRepository,
		logger:                 logger,
	}
}

func (c *OrganizationController) CreateOrg(ctx *gin.Context) {
	// Get current user ID from JWT token
	userID, _ := ctx.Get("user_id")

	// Retrieve user details from repository
	user, err := c.userRepository.GetUser(userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve user details"})
		return
	}

	// Parse request body
	var organization models.Organization
	if err := ctx.ShouldBindJSON(&organization); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add current user as the first member with access level 1
	member := models.OrganizationMember{
		Name:        user.Name,
		Email:       user.Email,
		AccessLevel: 1,
	}
	organization.OrganizationMembers = append(organization.OrganizationMembers, member)

	// Create organization
	organizationID, err := c.organizationRepository.CreateOrganization(&organization)
	if err != nil {
		if err.Error() == "organization name already exists" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "organization name already exists"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create organization"})
		}
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"organization_id": organizationID})
}

func (c *OrganizationController) GetOrgByID(ctx *gin.Context) {
	// Extract organization ID from the request URL
	orgID := ctx.Param("organization_id")

	// Retrieve organization details from repository
	org, err := c.organizationRepository.GetOrganizationByID(orgID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve organization"})
		return
	}

	// Check if the organization exists
	if org == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}

	// Get current user ID from JWT token
	userID, _ := ctx.Get("user_id")

	// Retrieve user details using the user ID
	user, err := c.userRepository.GetUser(userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve user details"})
		return
	}

	// Check if the user is a member of the organization
	accessLevel, err := c.organizationRepository.GetAccessLevelByEmail(orgID, user.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check access level"})
		return
	}

	if accessLevel == -1 {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "you are not a member of this organization"})
		return
	}

	// Return organization details
	ctx.JSON(http.StatusOK, org)
}

func (c *OrganizationController) UpdateOrg(ctx *gin.Context) {
	// Extract organization ID from the request URL
	orgID := ctx.Param("organization_id")

	// Retrieve organization details from repository
	org, err := c.organizationRepository.GetOrganizationByID(orgID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve organization"})
		return
	}

	// Check if the organization exists
	if org == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}

	// Get current user ID from JWT token
	userID, _ := ctx.Get("user_id")

	// Retrieve user details using the user ID
	user, err := c.userRepository.GetUser(userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve user details"})
		return
	}

	// Check if the user is a member of the organization with access level 1
	accessLevel, err := c.organizationRepository.GetAccessLevelByEmail(orgID, user.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check access level"})
		return
	}

	if accessLevel != 1 {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "you do not have sufficient access level to update this organization"})
		return
	}

	// Parse request body
	var updatedOrg models.Organization
	if err := ctx.ShouldBindJSON(&updatedOrg); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update organization details
	err = c.organizationRepository.UpdateOrganization(orgID, &updatedOrg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update organization"})
		return
	}

	// Return updated organization details
	ctx.JSON(http.StatusOK, updatedOrg)
}
func (c *OrganizationController) DeleteOrg(ctx *gin.Context) {
	// Extract organization ID from the request URL
	orgID := ctx.Param("organization_id")

	// Retrieve organization details from repository
	org, err := c.organizationRepository.GetOrganizationByID(orgID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve organization"})
		return
	}

	// Check if the organization exists
	if org == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}

	// Get current user ID from JWT token
	userID, _ := ctx.Get("user_id")

	// Retrieve user details using the user ID
	user, err := c.userRepository.GetUser(userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve user details"})
		return
	}

	// Check if the user is a member of the organization with access level 1
	accessLevel, err := c.organizationRepository.GetAccessLevelByEmail(orgID, user.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check access level"})
		return
	}

	if accessLevel != 1 {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "you do not have sufficient access level to delete this organization"})
		return
	}

	// Delete organization
	err = c.organizationRepository.DeleteOrganization(orgID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete organization"})
		return
	}

	// Return success message
	ctx.JSON(http.StatusOK, gin.H{"message": "organization deleted successfully"})
}

func (c *OrganizationController) InviteUser(ctx *gin.Context) {
	// Extract organization ID from the request URL
	orgID := ctx.Param("organization_id")

	// Retrieve organization details from repository
	org, err := c.organizationRepository.GetOrganizationByID(orgID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve organization"})
		return
	}

	// Check if the organization exists
	if org == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}

	// Get current user ID from JWT token
	userID, _ := ctx.Get("user_id")

	// Retrieve user details using the user ID
	user, err := c.userRepository.GetUser(userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve user details"})
		return
	}

	// Check if the user is a member of the organization with access level 1
	accessLevel, err := c.organizationRepository.GetAccessLevelByEmail(orgID, user.Email)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check access level"})
		return
	}

	if accessLevel != 1 {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "you do not have sufficient access level to invite users to this organization"})
		return
	}

	// Parse request body
	var inviteData struct {
		UserEmail string `json:"user_email" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&inviteData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Retrieve user details using the user ID
	invitee, err := c.userRepository.GetUserByEmail(inviteData.UserEmail)

	// Create the organization member object
	member := models.OrganizationMember{
		Name:        invitee.Name, // You may need to fill this with the invited user's name
		Email:       invitee.Email,
		AccessLevel: 0, // Set initial access level
	}

	// Add member to organization
	err = c.organizationRepository.AddMember(orgID, &member)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return success message
	ctx.JSON(http.StatusOK, gin.H{"message": "user invited to organization successfully"})
}
