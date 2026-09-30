package usecase

import (
	"context"
	"fmt"

	_ "embed"

	"github.com/Mpayy/e-commerce/services/order-service/internal/order/dto"
	"google.golang.org/genai"
)

//go:embed prompts/system_prompt.txt
var systemPromptText string

type AssistantUsecase interface {
	Chat(ctx context.Context, req dto.AssistantRequest) (*dto.AssistantResponse, error)
}

type AssistantUsecaseImpl struct {
	client       *genai.Client
	orderUsecase OrderUsecase
}

func NewAssistantUsecase(client *genai.Client, orderUsecase OrderUsecase) AssistantUsecase {
	return &AssistantUsecaseImpl{
		client:       client,
		orderUsecase: orderUsecase,
	}
}

func (a *AssistantUsecaseImpl) Chat(ctx context.Context, req dto.AssistantRequest) (*dto.AssistantResponse, error) {
	systemInstruction := &genai.Content{
		Parts: []*genai.Part{
			{Text: systemPromptText},
		},
	}

	tools := []*genai.Tool{
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{
					Name:        "get_sales_analytics",
					Description: "Mendapatkan laporan analitik penjualan dan performa produk berdasarkan rentang tanggal. Mengembalikan ringkasan akumulasi omset dan total order, rincian tren pendapatan harian (beserta running total), serta daftar produk terlaris (top products) berdasarkan jumlah terjual dan total pendapatan.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"from":  {Type: genai.TypeString, Description: "Tanggal awal filter analitik (format: YYYY-MM-DD). Jika tidak diisi, otomatis 30 hari sebelum tanggal 'to'."},
							"to":    {Type: genai.TypeString, Description: "Tanggal akhir filter analitik (format: YYYY-MM-DD). Jika tidak diisi, otomatis menggunakan tanggal hari ini."},
							"limit": {Type: genai.TypeInteger, Description: "Jumlah maksimal produk terlaris (top products) yang ingin ditampilkan (default: 5, maks: 50)."},
						},
					},
				},
				{
					Name:        "get_admin_order_list",
					Description: "Mengambil ringkasan daftar pesanan/order admin yang diurutkan dari yang terbaru. Mendukung penyaringan fleksibel berdasarkan status pesanan, ID pengguna (user_id), rentang nilai transaksi (min_amount & max_amount), serta rentang tanggal (from & to dengan format YYYY-MM-DD). Mengembalikan list order beserta metadata paginasi (total data dan total halaman).",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"status":     {Type: genai.TypeString, Description: "Filter berdasarkan status pesanan (contoh: PENDING, PAID, CANCELLED, COMPLETED)."},
							"user_id":    {Type: genai.TypeInteger, Description: "Filter berdasarkan ID pengguna/pembeli."},
							"min_amount": {Type: genai.TypeNumber, Description: "Nilai minimal total transaksi."},
							"max_amount": {Type: genai.TypeNumber, Description: "Nilai maksimal total transaksi."},
							"from":       {Type: genai.TypeString, Description: "Tanggal awal transaksi (format: YYYY-MM-DD)."},
							"to":         {Type: genai.TypeString, Description: "Tanggal akhir transaksi (format: YYYY-MM-DD)."},
							"page":       {Type: genai.TypeInteger, Description: "Halaman data yang ingin diambil (default: 1)."},
							"limit":      {Type: genai.TypeInteger, Description: "Jumlah data per halaman, maksimal 100 (default: 10)."},
						},
					},
				},
			},
		},
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: systemInstruction,
		Temperature:       genai.Ptr[float32](0),
		Tools:             tools,
	}

	contents := []*genai.Content{
		{Role: "user", Parts: []*genai.Part{{Text: req.Message}}},
	}

	const maxRounds = 5
	rawData := make([]any, 0)

	for range maxRounds {
		response, err := a.client.Models.GenerateContent(ctx, "gemini-3.1-flash-lite", contents, config)
		if err != nil {
			return nil, fmt.Errorf("failed generate content: %w", err)
		}

		if len(response.Candidates) == 0 || response.Candidates[0].Content == nil {
			return nil, fmt.Errorf("ai model no response")
		}

		var calls []*genai.FunctionCall
		for _, p := range response.Candidates[0].Content.Parts {
			if p.FunctionCall != nil {
				calls = append(calls, p.FunctionCall)
			}
		}

		if len(calls) == 0 {
			return &dto.AssistantResponse{Answer: response.Text(), RawData: rawData}, nil
		}

		responseParts := make([]*genai.Part, 0, len(calls))
		for _, call := range calls {
			var toolResult map[string]any
			switch call.Name {
			case "get_sales_analytics":
				result, err := a.orderUsecase.GetSalesAnalytics(ctx, &dto.SalesAnalyticsRequest{
					From:  argString(call.Args, "from"),
					To:    argString(call.Args, "to"),
					Limit: argInt(call.Args, "limit", 5),
				})
				if err != nil {
					toolResult = map[string]any{"error": err.Error()}
				} else {
					toolResult = map[string]any{"result": result}
					rawData = append(rawData, result)
				}

			case "get_admin_order_list":
				result, err := a.orderUsecase.GetAdminOrderList(ctx, &dto.AdminOrderListRequest{
					Status:    argString(call.Args, "status"),
					UserID:    argUint(call.Args, "user_id", uint(0)),
					MinAmount: argFloat(call.Args, "min_amount", float64(0)),
					MaxAmount: argFloat(call.Args, "max_amount", float64(0)),
					From:      argString(call.Args, "from"),
					To:        argString(call.Args, "to"),
					Page:      argInt(call.Args, "page", 1),
					Limit:     argInt(call.Args, "limit", 10),
				})
				if err != nil {
					toolResult = map[string]any{"error": err.Error()}
				} else {
					toolResult = map[string]any{"result": result}
					rawData = append(rawData, result)
				}

			default:
				toolResult = map[string]any{"error": fmt.Sprintf("unknown tool: %s", call.Name)}
			}

			responseParts = append(responseParts, &genai.Part{
				FunctionResponse: &genai.FunctionResponse{
					Name:     call.Name,
					Response: toolResult,
				},
			})
		}

		contents = append(contents,
			response.Candidates[0].Content,
			&genai.Content{Role: "user", Parts: responseParts},
		)
	}

	return nil, fmt.Errorf("ai tool loop exceeded")
}

func argString(args map[string]any, key string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return ""
}

func argInt(args map[string]any, key string, def int) int {
	if v, ok := args[key].(float64); ok { // JSON number selalu float64
		return int(v)
	}
	return def
}

func argFloat(args map[string]any, key string, def float64) float64 {
	if v, ok := args[key].(float64); ok {
		return v
	}
	return def
}

func argUint(args map[string]any, key string, def uint) uint {
	if v, ok := args[key].(float64); ok {
		return uint(v)
	}
	return def
}
