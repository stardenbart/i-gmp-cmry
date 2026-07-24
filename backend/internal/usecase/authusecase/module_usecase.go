package authusecase

import (
	"errors"

	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
)

type moduleUseCase struct {
	moduleRepo authdomain.ModuleRepository
}

func NewModuleUseCase(moduleRepo authdomain.ModuleRepository) authdomain.ModuleUseCase {
	return &moduleUseCase{moduleRepo: moduleRepo}
}

func (uc *moduleUseCase) GetAll() ([]authdomain.Module, error) {
	return uc.moduleRepo.FindAll()
}

func (uc *moduleUseCase) GetByID(id string) (*authdomain.Module, error) {
	return uc.moduleRepo.FindByID(id)
}

func (uc *moduleUseCase) Create(m *authdomain.Module) error {
	return uc.moduleRepo.Create(m)
}

func (uc *moduleUseCase) Update(m *authdomain.Module) error {
	return uc.moduleRepo.Update(m)
}

func (uc *moduleUseCase) Delete(id string) error {
	_, err := uc.moduleRepo.FindByID(id)
	if err != nil {
		return errors.New("module not found")
	}
	return uc.moduleRepo.Delete(id)
}
