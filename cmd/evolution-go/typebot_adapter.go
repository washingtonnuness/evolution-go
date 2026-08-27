package main

import (
	instance_model "github.com/evolution-foundation/evolution-go/pkg/instance/model"
	typebot_service "github.com/evolution-foundation/evolution-go/pkg/typebot/service"
	send_service "github.com/evolution-foundation/evolution-go/pkg/sendMessage/service"
)

// typebotSendAdapter adapta o send_service.SendService para a interface
// SendServiceInterface do pacote typebot, evitando import cycle.
type typebotSendAdapter struct {
	svc send_service.SendService
}

func newTypebotSendAdapter(svc send_service.SendService) *typebotSendAdapter {
	return &typebotSendAdapter{svc: svc}
}

func (a *typebotSendAdapter) SendText(text *typebot_service.TextStruct, instance *instance_model.Instance) error {
	_, err := a.svc.SendText(&send_service.TextStruct{
		Number: text.Number,
		Text:   text.Text,
	}, instance)
	return err
}

func (a *typebotSendAdapter) SendMediaUrl(media *typebot_service.MediaStruct, instance *instance_model.Instance) error {
	_, err := a.svc.SendMediaUrl(&send_service.MediaStruct{
		Number: media.Number,
		Type:   media.Type,
		Url:    media.Url,
	}, instance)
	return err
}

func (a *typebotSendAdapter) SendList(list *typebot_service.ListStruct, instance *instance_model.Instance) error {
	sections := make([]send_service.Section, 0, len(list.Sections))
	for _, sec := range list.Sections {
		rows := make([]send_service.Row, 0, len(sec.Rows))
		for _, r := range sec.Rows {
			rows = append(rows, send_service.Row{
				Title:       r.Title,
				Description: r.Description,
				RowId:       r.RowId,
			})
		}
		sections = append(sections, send_service.Section{
			Title: sec.Title,
			Rows:  rows,
		})
	}

	_, err := a.svc.SendList(&send_service.ListStruct{
		Number:      list.Number,
		Title:       list.Title,
		Description: list.Description,
		ButtonText:  list.ButtonText,
		FooterText:  list.FooterText,
		Sections:    sections,
		Delay:       int32(list.Delay),
	}, instance)
	return err
}

func (a *typebotSendAdapter) SendButton(button *typebot_service.ButtonStruct, instance *instance_model.Instance) error {
	btns := make([]send_service.Button, 0, len(button.Buttons))
	for _, b := range button.Buttons {
		btns = append(btns, send_service.Button{
			Type:        b.Type,
			DisplayText: b.DisplayText,
			Id:          b.Id,
			CopyCode:    b.CopyCode,
			URL:         b.URL,
			PhoneNumber: b.PhoneNumber,
			Currency:    b.Currency,
			Name:        b.Name,
			KeyType:     b.KeyType,
			Key:         b.Key,
		})
	}

	_, err := a.svc.SendButton(&send_service.ButtonStruct{
		Number:       button.Number,
		Title:        button.Title,
		Description:  button.Description,
		Footer:       button.Footer,
		Buttons:      btns,
	}, instance)
	return err
}
