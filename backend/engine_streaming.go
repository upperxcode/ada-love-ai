package backend

import (
	"context"
	"fmt"

	"reflect"

	"ada-love-ai/pkg/providers"
)

// StreamingWrapper envolve um LLMProvider para capturar deltas de streaming
// e emitir eventos de EventKindLLMDelta para o frontend do Ada-Love.
type StreamingWrapper struct {
	base     providers.LLMProvider
	eventBus *EventBus
}

func NewStreamingWrapper(base providers.LLMProvider) *StreamingWrapper {
	return &StreamingWrapper{
		base: base,
	}
}

func (w *StreamingWrapper) SetEventBus(eb *EventBus) {
	w.eventBus = eb
}

func (w *StreamingWrapper) Chat(ctx context.Context, messages []providers.Message, tools []providers.ToolDefinition, model string, options map[string]any) (*providers.LLMResponse, error) {
	// Usamos reflexão para detectar a assinatura de ChatStream (suporta 1 ou 2 argumentos no callback)
	v := reflect.ValueOf(w.base)
	m := v.MethodByName("ChatStream")

	if m.IsValid() {
		// Detecta o tipo do callback (último argumento).
		// NOTA: Para métodos, NumIn() inclui o receptor (w.base), então
		// ChatStream(ctx, msgs, tools, model, opts, cb) tem NumIn() == 7.
		mType := m.Type()
		fmt.Printf("[StreamingWrapper] Chat: ChatStream method found, NumIn=%d\n", mType.NumIn())
		if mType.NumIn() >= 6 {
			callbackType := mType.In(mType.NumIn() - 1)
			numArgs := callbackType.NumIn()
			fmt.Printf("[StreamingWrapper] Chat: callback type NumIn=%d\n", numArgs)

			var callbackValue reflect.Value
			lastLen := 0 // Track the last sent length for delta calculation
			if numArgs == 1 {
				cb := func(accumulated string) {
					if len(accumulated) > lastLen {
						delta := accumulated[lastLen:]
						lastLen = len(accumulated)
						w.emitDelta(ctx, delta)
					}
				}
				callbackValue = reflect.ValueOf(cb)
			} else if numArgs == 2 {
				cb := func(content string, reasoning string) {
					// Alguns provedores novos (como DeepSeek R1) enviam reasoning no segundo argumento
					if len(content) > lastLen {
						delta := content[lastLen:]
						lastLen = len(content)
						w.emitDelta(ctx, delta)
					}
				}
				callbackValue = reflect.ValueOf(cb)
			} else {
				fmt.Printf("[StreamingWrapper] Chat: callback has %d args, not 1 or 2, skipping streaming\n", numArgs)
			}

			if callbackValue.IsValid() {
				fmt.Printf("[StreamingWrapper] Chat: callback valid, calling ChatStream via reflection\n")
				results := m.Call([]reflect.Value{
					reflect.ValueOf(ctx),
					reflect.ValueOf(messages),
					reflect.ValueOf(tools),
					reflect.ValueOf(model),
					reflect.ValueOf(options),
					callbackValue,
				})

				if len(results) == 2 {
					res, _ := results[0].Interface().(*providers.LLMResponse)
					err, _ := results[1].Interface().(error)

					// Sincronização final: garante que qualquer conteúdo que não foi enviado via delta
					// durante o streaming seja enviado agora.
					if err == nil && res != nil && len(res.Content) > lastLen {
						w.emitDelta(ctx, res.Content[lastLen:])
					}

					return res, err
				}
			}
		}
	}

	// Fallback para chat normal se não suportar streaming ou reflexão falhar
	res, err := w.base.Chat(ctx, messages, tools, model, options)
	if err == nil && res != nil {
		w.emitDelta(ctx, res.Content)
	}
	return res, err
}

func (w *StreamingWrapper) ChatStream(ctx context.Context, messages []providers.Message, tools []providers.ToolDefinition, model string, options map[string]any, onChunk func(accumulated string)) (*providers.LLMResponse, error) {
	// Se a base já for um StreamingProvider, usamos diretamente
	if sp, ok := w.base.(providers.StreamingProvider); ok {
		lastLen := 0
		res, err := sp.ChatStream(ctx, messages, tools, model, options, func(accumulated string) {
			if len(accumulated) > lastLen {
				delta := accumulated[lastLen:]
				lastLen = len(accumulated)
				w.emitDelta(ctx, delta) // Notificamos nosso eventBus com o DELTA
			}
			if onChunk != nil {
				onChunk(accumulated)
			}
		})

		// Sincronização final
		if err == nil && res != nil && len(res.Content) > lastLen {
			w.emitDelta(ctx, res.Content[lastLen:])
		}
		return res, err
	}

	// Caso contrário, usamos nosso mecanismo de reflexão (que já tenta usar ChatStream se existir via nome)
	// chamando w.Chat, que já faz o wrap do callback e emite deltas.
	return w.Chat(ctx, messages, tools, model, options)
}

func (w *StreamingWrapper) GetDefaultModel() string {
	return w.base.GetDefaultModel()
}

func (w *StreamingWrapper) emitDelta(ctx context.Context, content string) {
	if w.eventBus == nil || content == "" {
		return
	}

	sessionID, _ := ctx.Value("session_id").(string)
	fmt.Printf("[StreamingWrapper] emitDelta: sessionID=%q content_len=%d\n", sessionID, len(content))
	w.eventBus.Emit(Event{Kind: EventKindLLMDelta, SessionID: sessionID, Payload: StreamingDeltaPayload{Content: content}})
}

func (e *Engine) injectEventBusToProvider(provider providers.LLMProvider) {
	wrapper, ok := provider.(*StreamingWrapper)
	if !ok {
		return
	}
	wrapper.SetEventBus(e.eventBus)
}
