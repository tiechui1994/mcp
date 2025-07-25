package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// handleDummyTool is a simple tool that returns "foo bar"
func handleCommandTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cmd, err := request.RequireString("cmd")
	if err != nil {
		return nil, err
	}

	cmd = strings.TrimSpace(cmd)
	if len(cmd) == 0 {
		return mcp.NewToolResultErrorf("invalid command"), nil
	}

	space := regexp.MustCompile(`\s+`)
	tokens := space.Split(cmd, -1)

	var command *exec.Cmd
	if len(tokens) == 1 {
		command = exec.CommandContext(ctx, tokens[0])
	} else {
		command = exec.CommandContext(ctx, tokens[0], tokens[1:]...)
	}

	var out bytes.Buffer
	command.Stderr = &out
	command.Stdout = &out
	command.Env = os.Environ()
	err = command.Run()
	if err != nil {
		return mcp.NewToolResultErrorf("exec %v", err), nil
	}

	return mcp.NewToolResultText(out.String()), nil
}

func handleRequestTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	u, err := request.RequireString("url")
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(u, "http") {
		return mcp.NewToolResultErrorf("url must http url"), nil
	}

	response, err := http.Get(u)
	if err != nil {
		return mcp.NewToolResultErrorf("http get %v", err), nil
	}

	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return mcp.NewToolResultErrorf("read response %v", err), nil
	}

	return mcp.NewToolResultText(string(raw)), nil
}

func NewMCPServer() *server.MCPServer {
	mcpServer := server.NewMCPServer(
		"mcp-server",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithToolCapabilities(true),
	)

	mcpServer.AddTool(mcp.NewTool("cmd",
		mcp.WithDescription("execute cmd tool"),
		mcp.WithString("cmd", mcp.Required()),
	), handleCommandTool)
	mcpServer.AddTool(mcp.NewTool("fetch",
		mcp.WithDescription("request url tool"),
		mcp.WithString("url", mcp.Required())), handleRequestTool)

	return mcpServer
}

var (
	port    int
	address string
)

func main() {
	flag.IntVar(&port, "port", 8000, "sse server port")
	flag.StringVar(&address, "ip", "localhost", "server ip address")
	flag.Parse()

	srv := server.NewStreamableHTTPServer(NewMCPServer(), server.WithHeartbeatInterval(time.Second*30))
	go func() {
		addr := fmt.Sprintf("%v:%v", address, port)
		if err := srv.Start(addr); err != nil {
			log.Fatalf("start error: %v", err)
		}
	}()

	sigtermC := make(chan os.Signal, 1)
	signal.Notify(sigtermC, os.Interrupt, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGKILL)

	<-sigtermC // block until SIGTERM is received
	log.Printf("SIGTERM received: gracefully shutting down...")
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
}
