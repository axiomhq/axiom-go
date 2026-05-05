// The purpose of this example is to show how to build a Google Agent
// Development Kit (ADK) agent with custom tools that interact with the Axiom
// API and the local filesystem. OpenTelemetry traces and logs from the ADK and
// the Axiom Go client are exported to Axiom.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/adk/telemetry"
	"google.golang.org/genai"

	"github.com/axiomhq/axiom-go/axiom"
	axiotel "github.com/axiomhq/axiom-go/axiom/otel"
)

const appName = "axiom_analyst"

func main() {
	// Export "AXIOM_TRACES_DATASET", "AXIOM_LOGS_DATASET" and "GOOGLE_API_KEY"
	// in addition to the required environment variables.

	tracesDataset := os.Getenv("AXIOM_TRACES_DATASET")
	if tracesDataset == "" {
		log.Fatal("AXIOM_TRACES_DATASET is required")
	}

	logsDataset := os.Getenv("AXIOM_LOGS_DATASET")
	if logsDataset == "" {
		log.Fatal("AXIOM_LOGS_DATASET is required")
	}

	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_API_KEY is required")
	}

	ctx := axiotel.WithCapability(context.Background(), appName)

	// 1. Create a sandboxed filesystem in a dedicated temp directory. Deferred
	// first so it closes last, after telemetry flushes.
	workDir, err := os.MkdirTemp("", "adk-example-*")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if removeErr := os.RemoveAll(workDir); removeErr != nil {
			log.Printf("removing work dir: %v", removeErr)
		}
	}()

	root, err := os.OpenRoot(workDir)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if closeErr := root.Close(); closeErr != nil {
			log.Printf("closing root: %v", closeErr)
		}
	}()

	// 2. Initialize OpenTelemetry providers for traces and logs, then register
	// them globally via ADK telemetry. The indirection through telemetry.New
	// and SetGlobalOtelProviders is required because
	// SetGenAICaptureMessageContent is in ADK's internal package.
	tp, err := axiotel.TracerProvider(ctx, tracesDataset, appName, "v1.0.0")
	if err != nil {
		log.Fatal(err)
	}

	lp, err := axiotel.LoggerProvider(ctx, logsDataset, appName, "v1.0.0")
	if err != nil {
		log.Fatal(err)
	}

	providers, err := telemetry.New(ctx,
		telemetry.WithTracerProvider(tp),
		telemetry.WithLoggerProvider(lp),
		telemetry.WithGenAICaptureMessageContent(true),
	)
	if err != nil {
		log.Fatal(err)
	}
	providers.SetGlobalOtelProviders()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
		defer cancel()
		if err := providers.Shutdown(shutdownCtx); err != nil {
			log.Printf("telemetry shutdown: %v", err)
		}
	}()

	// 3. Initialize the Axiom API client.
	client, err := axiom.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	// 4. Create tools.
	tools, err := createTools(client.Datasets, root)
	if err != nil {
		log.Fatal(err)
	}

	// 5. Create agent with a Gemini model.
	model, err := gemini.NewModel(ctx, "gemini-3.1-pro-preview", &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatal(err)
	}

	analystAgent, err := llmagent.New(llmagent.Config{
		Name:        appName,
		Model:       model,
		Description: "Agent that analyzes Axiom datasets and writes reports",
		Instruction: systemInstruction,
		Tools:       tools,
	})
	if err != nil {
		log.Fatal(err)
	}

	// 6. Create runner and session.
	sessionService := session.InMemoryService()
	r, err := runner.New(runner.Config{
		AppName:        appName,
		Agent:          analystAgent,
		SessionService: sessionService,
	})
	if err != nil {
		log.Fatal(err)
	}

	sess, err := sessionService.Create(ctx, &session.CreateRequest{
		AppName: appName,
		UserID:  "user",
	})
	if err != nil {
		log.Fatal(err)
	}

	// 7. Run a multi-turn conversation.
	sessionID := sess.Session.ID()

	turns := []struct {
		step string
		text string
	}{
		{"analyze", "Analyze my datasets. For each dataset, " +
			"query the last 24 hours to get event counts and a sample of recent " +
			"events. Write a comprehensive summary report to report.md that " +
			"includes statistics and notable content from each dataset."},
		{"cleanup", "Delete the report and confirm what you removed."},
	}

	for i, turn := range turns {
		stepCtx := axiotel.WithStep(ctx, turn.step)
		fmt.Printf("\n--- turn %d: %s ---\n", i+1, turn.step)

		userMessage := genai.NewContentFromText(turn.text, "user")
		for event, err := range r.Run(stepCtx, "user", sessionID, userMessage, agent.RunConfig{}) {
			if err != nil {
				log.Fatal(err)
			}
			if event.Content == nil {
				continue
			}
			for _, part := range event.Content.Parts {
				switch {
				case part.FunctionCall != nil:
					fmt.Printf("=> tool: %s(%s)\n", part.FunctionCall.Name, formatArgs(part.FunctionCall.Args))
				case part.Text != "":
					fmt.Println(part.Text)
				}
			}
		}
	}
}

const systemInstruction = `You are an Axiom data analyst agent. You can query
observability datasets, persist your work to the filesystem, and fetch external
resources.

When analyzing datasets, use APL queries to gather both statistics and sample
content. Write your findings to files in markdown format. Structure reports with
clear headings, tables, and summaries.`

func formatArgs(args map[string]any) string {
	parts := make([]string, 0, len(args))
	for k, v := range args {
		s := fmt.Sprintf("%v", v)
		if len(s) > 60 {
			s = s[:57] + "..."
		}
		parts = append(parts, fmt.Sprintf("%s=%q", k, s))
	}
	return strings.Join(parts, ", ")
}
