package user_repository

import "testing"

func TestCreateAndFind(t *testing.T) {
	repo := NewUserRepository()

	if _, err := repo.Create("budi", "Budi Santoso", "hash1"); err != nil {
		t.Fatalf("create: %v", err)
	}

	user, err := repo.FindByUsername("budi")
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if user.FullName != "Budi Santoso" || user.PasswordHash != "hash1" {
		t.Errorf("data pengguna tidak sesuai: %+v", user)
	}
}

func TestDuplicateUsername(t *testing.T) {
	repo := NewUserRepository()

	if _, err := repo.Create("budi", "Budi", "h"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := repo.Create("budi", "Budi Lain", "h2"); err != ErrUsernameTaken {
		t.Errorf("expected ErrUsernameTaken, got %v", err)
	}
}

func TestFindNotFound(t *testing.T) {
	repo := NewUserRepository()

	if _, err := repo.FindByUsername("tidakada"); err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUpdatePassword(t *testing.T) {
	repo := NewUserRepository()

	if _, err := repo.Create("budi", "Budi", "old"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := repo.UpdatePassword("budi", "new"); err != nil {
		t.Fatalf("update: %v", err)
	}

	user, err := repo.FindByUsername("budi")
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if user.PasswordHash != "new" {
		t.Errorf("password hash tidak berubah: %s", user.PasswordHash)
	}

	if err := repo.UpdatePassword("hantu", "x"); err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}
