package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/suntao12138/pikpakdriver/mcp/server"
	"github.com/suntao12138/pikpakdriver/pkg/pikpak"
)

var (
	email    = flag.String("email", "", "PikPak account email (for login)")
	password = flag.String("password", "", "PikPak account password (for login)")
	proxy    = flag.String("proxy", "", "HTTP proxy URL (e.g. http://127.0.0.1:7890)")
	help     = flag.Bool("help", false, "display help information")
)

func main() {
	flag.Parse()

	if *help {
		printUsage()
		os.Exit(1)
	}

	// ── Login Mode ─────────────────────────────────────────────────────
	if *email != "" && *password != "" {
		if err := doLoginAndSave(*email, *password, *proxy); err != nil {
			log.Fatalf("Login failed: %v", err)
		}
		log.Printf("Login successful! Credentials saved to %s", pikpak.CredentialsPath())
		log.Printf("Session saved to %s", pikpak.SessionPath())
		if *proxy != "" {
			log.Printf("Proxy setting saved: %s", *proxy)
		}
		log.Printf("Next time just run '%s' — it will auto-login from saved credentials.", os.Args[0])
		return
	}

	// ── MCP Server Mode ────────────────────────────────────────────────
	if err := pikpak.EnsureDir(); err != nil {
		log.Fatalf("Cannot create config dir: %v", err)
	}

	// NewClient handles proxy priority: --proxy flag > config.json > none
	client, err := pikpak.NewClient(*proxy)
	if err != nil {
		log.Fatalf("Failed to create PikPak client: %v\nRun: %s --email your@email --password yourpass", err, os.Args[0])
	}

	mcpServer := server.NewServer().WithClient(client)
	log.Printf("Starting pikpakdriver MCP server (session: %s)", pikpak.SessionPath())
	if err := mcpServer.Start(context.Background()); err != nil {
		log.Fatalf("Server exited: %v", err)
	}
}

func doLoginAndSave(email, password, proxyURL string) error {
	if err := pikpak.EnsureDir(); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	// Save credentials first (including proxy setting)
	if err := pikpak.SaveCredentials(email, password, proxyURL); err != nil {
		return fmt.Errorf("save credentials: %w", err)
	}

	client := pikpak.NewLoginClient(email, proxyURL)
	_, err := client.Login(email, password)
	if err != nil {
		// Login failed — remove saved credentials to avoid repeated failures
		os.Remove(pikpak.CredentialsPath())

		// Check for region restriction first
		var regionErr *pikpak.RegionError
		if errors.As(err, &regionErr) {
			fmt.Fprintf(os.Stderr, "\n")
			fmt.Fprintf(os.Stderr, "╔══════════════════════════════════════════════════════════════╗\n")
			fmt.Fprintf(os.Stderr, "║  ❌ 登录失败：PikPak 限制了您所在地区的访问               ║\n")
			fmt.Fprintf(os.Stderr, "║                                                              ║\n")
			fmt.Fprintf(os.Stderr, "║  PikPak 屏蔽中国大陆 IP 地址，您需要使用代理/VPN 后重试。    ║\n")
			fmt.Fprintf(os.Stderr, "║                                                              ║\n")
			fmt.Fprintf(os.Stderr, "║  解决方案：                                                    ║\n")
			fmt.Fprintf(os.Stderr, "║  1. 使用 --proxy 参数指定 HTTP 代理：                          ║\n")
			fmt.Fprintf(os.Stderr, "║     %s --email x --password x --proxy http://127.0.0.1:7890    ║\n", os.Args[0])
			fmt.Fprintf(os.Stderr, "║  2. 代理地址会被保存到配置文件中，后续不再需要 --proxy 参数   ║\n")
			fmt.Fprintf(os.Stderr, "╚══════════════════════════════════════════════════════════════╝\n")
			fmt.Fprintf(os.Stderr, "\n")
			return regionErr
		}

		return err
	}

	return nil
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "PikPak Driver MCP Server — independent from pikpaktui\n\n")
	fmt.Fprintf(os.Stderr, "Modes:\n")
	fmt.Fprintf(os.Stderr, "  1. First-time setup:  %s --email your@email --password yourpass [--proxy http://...]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "     (saves email/password/proxy to %s)\n", pikpak.CredentialsPath())
	fmt.Fprintf(os.Stderr, "  2. MCP Server:        %s [--proxy http://...]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "     (auto-login from saved credentials; --proxy overrides config)\n")
	fmt.Fprintf(os.Stderr, "\nProxy priority: --proxy CLI flag  >  config.json proxy  >  no proxy\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nConfig directory: %s\n", pikpak.ConfigDir())
	fmt.Fprintf(os.Stderr, "  Credentials: %s\n", pikpak.CredentialsPath())
	fmt.Fprintf(os.Stderr, "  Session:     %s\n", pikpak.SessionPath())
}
