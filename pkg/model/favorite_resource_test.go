package model

import "testing"

func TestDesktopFavoriteResourceCRUD(t *testing.T) {
	if err := DB.Exec("DELETE FROM favorite_resources").Error; err != nil {
		t.Fatalf("cleanup favorite_resources failed: %v", err)
	}
	if err := DB.Where("username = ?", LocalDesktopUser.Username).Delete(&User{}).Error; err != nil {
		t.Fatalf("cleanup local desktop user failed: %v", err)
	}

	first, err := SaveDesktopFavoriteResource("cluster-a", "Deployments", "default", "nginx")
	if err != nil {
		t.Fatalf("SaveDesktopFavoriteResource() first error = %v", err)
	}
	if first == nil || first.ID == 0 {
		t.Fatalf("SaveDesktopFavoriteResource() first = %#v, want persisted favorite", first)
	}

	second, err := SaveDesktopFavoriteResource("cluster-a", "deployments", "default", "nginx")
	if err != nil {
		t.Fatalf("SaveDesktopFavoriteResource() second error = %v", err)
	}
	if second == nil || second.ID != first.ID {
		t.Fatalf("duplicate favorite created new row: first=%#v second=%#v", first, second)
	}

	otherCluster, err := SaveDesktopFavoriteResource("cluster-b", "deployments", "default", "nginx")
	if err != nil {
		t.Fatalf("SaveDesktopFavoriteResource() other cluster error = %v", err)
	}
	if otherCluster == nil || otherCluster.ID == first.ID {
		t.Fatalf("other cluster favorite should be separate row: first=%#v other=%#v", first, otherCluster)
	}

	clusterAFavorites, err := ListDesktopFavoriteResources("cluster-a")
	if err != nil {
		t.Fatalf("ListDesktopFavoriteResources(cluster-a) error = %v", err)
	}
	if len(clusterAFavorites) != 1 {
		t.Fatalf("ListDesktopFavoriteResources(cluster-a) len = %d, want 1", len(clusterAFavorites))
	}
	if clusterAFavorites[0].ResourceType != "deployments" {
		t.Fatalf("ResourceType = %q, want %q", clusterAFavorites[0].ResourceType, "deployments")
	}

	clusterBFavorites, err := ListDesktopFavoriteResources("cluster-b")
	if err != nil {
		t.Fatalf("ListDesktopFavoriteResources(cluster-b) error = %v", err)
	}
	if len(clusterBFavorites) != 1 {
		t.Fatalf("ListDesktopFavoriteResources(cluster-b) len = %d, want 1", len(clusterBFavorites))
	}

	if err := DeleteDesktopFavoriteResource("cluster-a", "deployments", "default", "nginx"); err != nil {
		t.Fatalf("DeleteDesktopFavoriteResource() error = %v", err)
	}

	clusterAFavorites, err = ListDesktopFavoriteResources("cluster-a")
	if err != nil {
		t.Fatalf("ListDesktopFavoriteResources(cluster-a) after delete error = %v", err)
	}
	if len(clusterAFavorites) != 0 {
		t.Fatalf("ListDesktopFavoriteResources(cluster-a) after delete len = %d, want 0", len(clusterAFavorites))
	}
}
