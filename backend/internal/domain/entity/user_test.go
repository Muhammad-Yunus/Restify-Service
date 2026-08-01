package entity

import "testing"

func TestUserSetPassword(t *testing.T) {
	u := &User{}

	if err := u.SetPassword("s3cret-pass"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}

	if u.PasswordHash == "" {
		t.Fatal("password hash is empty")
	}

	if u.PasswordHash == "s3cret-pass" {
		t.Fatal("password stored in plaintext")
	}
}

func TestUserCheckPassword(t *testing.T) {
	u := &User{}

	if err := u.SetPassword("s3cret-pass"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}

	if !u.CheckPassword("s3cret-pass") {
		t.Error("CheckPassword returned false for correct password")
	}

	if u.CheckPassword("wrong-password") {
		t.Error("CheckPassword returned true for wrong password")
	}
}

func TestUserHasRole(t *testing.T) {
	u := &User{Roles: []*Role{{Name: RoleDeveloper}}}

	if !u.HasRole(RoleDeveloper) {
		t.Error("HasRole returned false for existing role")
	}

	if u.HasRole(RoleAdministrator) {
		t.Error("HasRole returned true for missing role")
	}
}

func TestUserValidate(t *testing.T) {
	tests := []struct {
		name string
		user User
		want bool
	}{
		{name: "valid", user: User{Email: "user@example.com", PasswordHash: "hash"}, want: true},
		{name: "empty email", user: User{PasswordHash: "hash"}, want: false},
		{name: "invalid email format", user: User{Email: "not-an-email", PasswordHash: "hash"}, want: false},
		{name: "missing password hash", user: User{Email: "user@example.com"}, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.user.Validate()
			if (err == nil) != tc.want {
				t.Errorf("Validate() error = %v, want success = %v", err, tc.want)
			}
		})
	}
}

func TestEmailValid(t *testing.T) {
	if !EmailValid("user@example.com") {
		t.Error("EmailValid returned false for valid email")
	}

	if EmailValid("not-an-email") {
		t.Error("EmailValid returned true for invalid email")
	}
}
