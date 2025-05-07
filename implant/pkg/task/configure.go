package task

type Configure struct {
	NewJitter         float64
	NewDwell          string
	NewC2CommProtocol string
	ID                uint `json:"id"`
}

func NewConfigureImplantTask(ID uint, config map[string]interface{}) *Configure {

	configureImplant := new(Configure)
	if raw, ok := config["NewDwell"]; ok {
		if s, ok := raw.(string); ok {
			configureImplant.NewDwell = s
		}
	}
	if raw, ok := config["NewJitter"]; ok {
		if f, ok := raw.(float64); ok {
			configureImplant.NewJitter = f
		}
	}
	if raw, ok := config["NewC2CommProtocol"]; ok {
		if s, ok := raw.(string); ok {
			configureImplant.NewC2CommProtocol = s
		}
	}
	configureImplant.ID = ID
	return configureImplant
}

// THIS WILL NOT BE CALLED AS CONFIGUREIMPLANT IS A SPECIAL TASK
func (p *Configure) Do() Result {
	return *new(Result)
}

func (p *Configure) GetID() uint {
	return p.ID
}
