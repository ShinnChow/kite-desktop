package model

import (
	"errors"
	"strings"
)

type FavoriteResource struct {
	Model
	UserID uint `json:"userId" gorm:"not null;index:idx_favorite_resources_lookup,priority:1;uniqueIndex:idx_favorite_resource_identity,priority:1"`

	ClusterName  string `json:"clusterName" gorm:"type:varchar(100);not null;index:idx_favorite_resources_lookup,priority:2;uniqueIndex:idx_favorite_resource_identity,priority:2"`
	ResourceType string `json:"resourceType" gorm:"type:varchar(50);not null;index:idx_favorite_resources_lookup,priority:3;uniqueIndex:idx_favorite_resource_identity,priority:3"`
	Namespace    string `json:"namespace,omitempty" gorm:"type:varchar(100);uniqueIndex:idx_favorite_resource_identity,priority:4"`
	ResourceName string `json:"resourceName" gorm:"type:varchar(255);not null;index:idx_favorite_resources_lookup,priority:4;uniqueIndex:idx_favorite_resource_identity,priority:5"`

	User *User `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func normalizeFavoriteResourceInput(clusterName, resourceType, namespace, resourceName string) (string, string, string, string, error) {
	normalizedCluster := strings.TrimSpace(clusterName)
	if normalizedCluster == "" {
		return "", "", "", "", errors.New("cluster name is required")
	}

	normalizedResourceType := strings.ToLower(strings.TrimSpace(resourceType))
	normalizedNamespace := strings.TrimSpace(namespace)
	normalizedResourceName := strings.TrimSpace(resourceName)
	if normalizedResourceType == "" {
		return "", "", "", "", errors.New("resource type is required")
	}
	if normalizedResourceName == "" {
		return "", "", "", "", errors.New("resource name is required")
	}

	return normalizedCluster, normalizedResourceType, normalizedNamespace, normalizedResourceName, nil
}

func ListDesktopFavoriteResources(clusterName string) ([]FavoriteResource, error) {
	if DB == nil {
		return nil, nil
	}

	normalizedCluster := strings.TrimSpace(clusterName)
	if normalizedCluster == "" {
		return []FavoriteResource{}, nil
	}

	user, err := EnsureLocalDesktopUser()
	if err != nil {
		return nil, err
	}

	var favorites []FavoriteResource
	err = DB.
		Where("user_id = ? AND cluster_name = ?", user.ID, normalizedCluster).
		Order("created_at DESC").
		Find(&favorites).Error
	return favorites, err
}

func SaveDesktopFavoriteResource(clusterName, resourceType, namespace, resourceName string) (*FavoriteResource, error) {
	if DB == nil {
		return nil, nil
	}

	normalizedCluster, normalizedResourceType, normalizedNamespace, normalizedResourceName, err := normalizeFavoriteResourceInput(clusterName, resourceType, namespace, resourceName)
	if err != nil {
		return nil, err
	}

	user, err := EnsureLocalDesktopUser()
	if err != nil {
		return nil, err
	}

	favorite := FavoriteResource{
		UserID:       user.ID,
		ClusterName:  normalizedCluster,
		ResourceType: normalizedResourceType,
		Namespace:    normalizedNamespace,
		ResourceName: normalizedResourceName,
	}

	err = DB.
		Where(&FavoriteResource{
			UserID:       favorite.UserID,
			ClusterName:  favorite.ClusterName,
			ResourceType: favorite.ResourceType,
			Namespace:    favorite.Namespace,
			ResourceName: favorite.ResourceName,
		}).
		FirstOrCreate(&favorite).Error
	if err != nil {
		return nil, err
	}

	return &favorite, nil
}

func DeleteDesktopFavoriteResource(clusterName, resourceType, namespace, resourceName string) error {
	if DB == nil {
		return nil
	}

	normalizedCluster, normalizedResourceType, normalizedNamespace, normalizedResourceName, err := normalizeFavoriteResourceInput(clusterName, resourceType, namespace, resourceName)
	if err != nil {
		return err
	}

	user, err := EnsureLocalDesktopUser()
	if err != nil {
		return err
	}

	return DB.
		Where(
			"user_id = ? AND cluster_name = ? AND resource_type = ? AND namespace = ? AND resource_name = ?",
			user.ID,
			normalizedCluster,
			normalizedResourceType,
			normalizedNamespace,
			normalizedResourceName,
		).
		Delete(&FavoriteResource{}).Error
}
