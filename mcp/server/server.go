package server

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/suntao12138/pikpakdriver/mcp/server/tools"
	"github.com/suntao12138/pikpakdriver/pkg/pikpak"
)

type Server struct {
	mcpServer *mcp.Server
	client    *pikpak.Client
}

func NewServer() *Server {
	return &Server{
		mcpServer: mcp.NewServer(&mcp.Implementation{
			Name:    "pikpakdriver-mcp-server",
			Version: "1.0.0",
		}, nil),
	}
}

func (s *Server) WithClient(client *pikpak.Client) *Server {
	s.client = client
	return s
}

func (s *Server) Start(ctx context.Context) error {
	s.registerTools()
	if err := s.mcpServer.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Printf("Server failed: %v", err)
		return err
	}
	return nil
}

func (s *Server) registerTools() {
	tools.NewAccountTools(s.client).RegisterTools(s.mcpServer)
	tools.NewFileTools(s.client).RegisterTools(s.mcpServer)
	tools.NewOfflineTools(s.client).RegisterTools(s.mcpServer)
	tools.NewShareTools(s.client).RegisterTools(s.mcpServer)
	tools.NewEventsTools(s.client).RegisterTools(s.mcpServer)
}
