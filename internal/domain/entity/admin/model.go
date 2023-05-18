package admin

type Admin struct {
	UserID   string `json:"user_id"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type DTO struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (a *Admin) ToDTO() *DTO {
	return &DTO{
		Name:     a.Name,
		Password: a.Password,
	}
}

func (dto *DTO) ToDomain() *Admin {
	return &Admin{
		UserID:   "",
		Name:     dto.Name,
		Password: dto.Password,
	}
}
