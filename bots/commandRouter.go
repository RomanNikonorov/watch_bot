package bots

import (
	"log"
	"strings"
)

// CommandHandler interface for command implementations
type CommandHandler interface {
	Execute(cmd Command) (string, error)
	Description() string
}

// CommandRouter routes commands to appropriate handlers
type CommandRouter struct {
	handlers map[string]CommandHandler
}

// NewCommandRouter creates a new command router
func NewCommandRouter() *CommandRouter {
	return &CommandRouter{
		handlers: make(map[string]CommandHandler),
	}
}

// Register adds a command handler to the router
func (r *CommandRouter) Register(name string, handler CommandHandler) {
	r.handlers[name] = handler
}

// GetRegisteredCommands returns list of registered command names with descriptions
func (r *CommandRouter) GetRegisteredCommands() map[string]string {
	commands := make(map[string]string)
	for name, handler := range r.handlers {
		commands[name] = handler.Description()
	}
	return commands
}

// Handle processes a command and returns response
func (r *CommandRouter) Handle(cmd Command) (string, error) {
	handler, exists := r.handlers[cmd.Name]
	if !exists {
		return r.unknownCommandResponse(), nil
	}
	return handler.Execute(cmd)
}

func (r *CommandRouter) unknownCommandResponse() string {
	var sb strings.Builder
	sb.WriteString("Unknown command. Available commands:\n")
	for name, desc := range r.GetRegisteredCommands() {
		sb.WriteString("\\" + name + " - " + desc + "\n")
	}
	return sb.String()
}

// Listen starts listening for commands and sends responses
func (r *CommandRouter) Listen(commandChannel chan Command, messagesChannel chan Message) {
	for cmd := range commandChannel {
		response, err := r.Handle(cmd)
		if err != nil {
			log.Printf("Error handling command %s: %v", cmd.Name, err)
			response = "An error occurred while executing the command"
		}
		messagesChannel <- Message{
			ChatId: cmd.ChatId,
			Text:   response,
		}
	}
}

// ParseCommand parses message text into Command struct
// Returns nil if message is not a command (doesn't start with \)
func ParseCommand(text string, chatId string) *Command {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, "\\") {
		return nil
	}

	// Remove prefix and split into parts
	text = strings.TrimPrefix(text, "\\")
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return nil
	}

	cmd := &Command{
		Name:   strings.ToLower(parts[0]),
		ChatId: chatId,
		Params: make(map[string]string),
	}

	// Parse remaining parts as positional params
	for i, part := range parts[1:] {
		cmd.Params[string(rune('0'+i))] = part
	}

	return cmd
}
