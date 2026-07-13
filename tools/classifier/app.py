from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import List
import os
import logging
from huggingface_hub import hf_hub_download
from llama_cpp import Llama

# Configuração do Logger
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s | %(levelname)-8s | %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S"
)
logger = logging.getLogger("classifier")

app = FastAPI(title="Local GGUF Intent Classifier")

class ClassifyReq(BaseModel):
    text: str
    candidate_labels: List[str]

class ClassifyResp(BaseModel):
    labels: List[str]
    scores: List[float]
    top_label: str

_llm = None

def get_llm():
    global _llm
    if _llm is None:
        try:
            logger.info("🚀 Baixando/Verificando modelo GGUF estável...")
            # Baixa o Qwen 2.5 1.5B Instruct (Formato GGUF, Q4_K_M é o balanço perfeito)
            model_path = hf_hub_download(
                repo_id="Qwen/Qwen2.5-1.5B-Instruct-GGUF",
                filename="qwen2.5-1.5b-instruct-q4_k_m.gguf"
            )

            logger.info(f"📦 Carregando GGUF na memória: {model_path}")
            # Inicializa o motor do llama.cpp
            _llm = Llama(
                model_path=model_path,
                n_ctx=1024,       # Contexto curto para router (economiza muita RAM)
                n_threads=4,      # Ajuste conforme seus cores de CPU
                verbose=False     # Desliga os logs gigantes de C++ no meio do seu log
            )
            logger.info("✅ GGUF carregado e pronto para inferência!")
        except Exception as e:
            logger.error(f"❌ Falha ao inicializar o motor GGUF: {e}")
            raise RuntimeError(f"failed to initialize GGUF engine: {e}")
    return _llm

@app.post("/classify", response_model=ClassifyResp)
async def classify(req: ClassifyReq):
    if not req.candidate_labels or not req.text:
        raise HTTPException(status_code=400, detail="text and candidate_labels required")

    logger.info(f"📥 ENTRADA | Texto: '{req.text}'")

    llm = get_llm()

    # Como LLMs GGUF não têm uma função nativa de pipeline zero-shot,
    # nós usamos Engenharia de Prompt Estruturada para fazê-lo responder APENAS a label correta.
    labels_str = ", ".join([f"'{l}'" for l in req.candidate_labels])

    prompt = f"""<|im_start|>system
Você é um classificador especializado em engenharia de software e tecnologia.
Sua única tarefa é ler o texto ou snippet de código fornecido e identificar qual das categorias fornecidas melhor descreve o assunto principal.

Regras estritas:
1. Responda APENAS E ESTRITAMENTE com o nome exato da categoria escolhida, idêntica à lista fornecida.
2. Não adicione pontuação, explicações, aspas ou markdown.
3. Se o texto mencionar código ou conceitos de uma tecnologia específica (como Go, Flutter, React), escolha a respectiva tecnologia.<|im_end|>
<|im_start|>user
Opções válidas de categorias: [{labels_str}]

Texto para analisar: "{req.text}"

Qual a categoria correta?<|im_end|>
<|im_start|>assistant
"""

    try:
        # Inferência ultra rápida usando o modelo GGUF
        output = llm(
            prompt,
            max_tokens=32,
            temperature=0.0, # Temperatura 0 garante determinismo absoluto
            stop=["<|im_end|>", "\n"]
        )

        # Limpa a resposta do modelo
        predicted_label = output["choices"][0]["text"].strip().replace("'", "").replace('"', '')
        logger.info(f"🤖 Resposta bruta do modelo: '{predicted_label}'")

        # Valida se a resposta do modelo bate com as opções enviadas pelo Go
        if predicted_label in req.candidate_labels:
            top_label = predicted_label
        else:
            # Fallback caso o modelo invente uma label ou adicione ruído
            top_label = req.candidate_labels[0]
            for label in req.candidate_labels:
                if label.lower() in predicted_label.lower():
                    top_label = label
                    break

        # Como modelos de chat cospem texto e não probabilidades em array,
        # nós mockamos os scores para manter a compatibilidade com o seu código Go atual.
        scores = [1.0 if l == top_label else 0.0 for l in req.candidate_labels]

        logger.info(f"📤 SAÍDA   | Top Label: '{top_label}'")
        return ClassifyResp(labels=req.candidate_labels, scores=scores, top_label=top_label)

    except Exception as e:
        logger.error(f"❌ Erro durante a inferência GGUF: {e}")
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8008)
