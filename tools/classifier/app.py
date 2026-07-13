import logging
import os
from typing import List
from fastapi import FastAPI, HTTPException
from huggingface_hub import hf_hub_download
from llama_cpp import Llama
from pydantic import BaseModel

# Configuração do Logger para monitoramento no terminal
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s | %(levelname)-8s | %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S"
)
logger = logging.getLogger("classifier")

app = FastAPI(title="Classifier Nano Router GGUF")

# Pydantic Schemas para manter compatibilidade exata com o seu Backend em Go
class ClassifyReq(BaseModel):
    text: str
    candidate_labels: List[str] = [] # Deixado como opcional pois agora o Python gerencia as chaves internas

class ClassifyResp(BaseModel):
    labels: List[str]
    scores: List[float]
    top_label: str

# Mapeamento Estrito: Dá contexto semântico pesado para o Qwen 1.5B não errar,
# mas mapeia direto para as chaves curtas que o seu switch/case em Go precisa.
LABEL_MAP = {
    "desenvolvimento de software e linguagem go": "GO",
    "desenvolvimento de software e framework react": "REACT",
    "desenvolvimento de software e framework flutter": "FLUTTER",
    "assunto geral, perguntas gerais ou conversacao comum": "GERAL"
}

_llm = None

def get_llm():
    global _llm
    if _llm is None:
        try:
            logger.info("🚀 Baixando/Verificando modelo GGUF estável...")
            # Baixa e gerencia o cache do modelo Qwen 2.5 1.5B Instruct
            model_path = hf_hub_download(
                repo_id="Qwen/Qwen2.5-1.5B-Instruct-GGUF",
                filename="qwen2.5-1.5b-instruct-q4_k_m.gguf"
            )

            logger.info(f"📦 Carregando GGUF na memória: {model_path}")
            # Inicializa os bindings C++ do llama.cpp
            _llm = Llama(
                model_path=model_path,
                n_ctx=1024,       # Contexto curto otimizado para roteador (consome menos RAM)
                n_threads=4,      # Ajustado para os cores da sua CPU
                verbose=False     # Silencia o dump de logs brutos de C++ no seu terminal
            )
            logger.info("✅ GGUF carregado e pronto para inferência!")
        except Exception as e:
            logger.error(f"❌ Falha ao inicializar o motor GGUF: {e}")
            raise RuntimeError(f"failed to initialize GGUF engine: {e}")
    return _llm

@app.post("/classify", response_model=ClassifyResp)
async def classify(req: ClassifyReq):
    if not req.text:
        raise HTTPException(status_code=400, detail="text required")

    chaves_curtas = list(LABEL_MAP.values())
    texto_limpo = req.text.strip().lower()

    # ⚡ FILTRO 1: Intercepta o Health Check do Orquestrador sem gastar processamento de IA
    if "health check" in texto_limpo or texto_limpo == "ping":
        logger.info("📡 [Health Check] Ignorando inferência e retornando GERAL")
        scores = [1.0 if c == "GERAL" else 0.0 for c in chaves_curtas]
        return ClassifyResp(labels=chaves_curtas, scores=scores, top_label="GERAL")

    logger.info(f"📥 ENTRADA | Texto: '{req.text}'")

    # Lista de frases longas que serão injetadas no prompt da IA
    candidate_labels = list(LABEL_MAP.keys())
    labels_str = ", ".join([f"'{l}'" for l in candidate_labels])

    # Prompt estruturado ChatML para o Qwen atuar como classificador rígido
    prompt = f"""<|im_start|>system
Você é um classificador especializado em engenharia de software e tecnologia.
Sua única tarefa é ler o texto do usuário e escolher a categoria que melhor se aplica dentre as opções fornecidas.
Responda APENAS E ESTRITAMENTE com o nome exato da categoria escolhida, sem pontuação, sem explicações e sem markdown.<|im_end|>
<|im_start|>user
Opções válidas de categorias: [{labels_str}]

Texto para analisar: "{req.text}"

Qual a categoria correta?<|im_end|>
<|im_start|>assistant
"""

    try:
        llm = get_llm()

        # Inferência local determinística via llama.cpp
        output = llm(
            prompt,
            max_tokens=64,
            temperature=0.0, # Temperatura zero garante que a resposta não mude entre chamadas
            stop=["<|im_end|>", "\n"]
        )

        predicted_label = output["choices"][0]["text"].strip().replace("'", "").replace('"', '')
        logger.info(f"🤖 Resposta bruta do modelo: '{predicted_label}'")

        # Faz o cruzamento inteligente da resposta da IA com as chaves descritivas do dicionário
        matched_desc_label = candidate_labels[-1] # Fallback padrão inicial (Assunto Geral)
        for label in candidate_labels:
            if label.lower() in predicted_label.lower() or predicted_label.lower() in label.lower():
                matched_desc_label = label
                break

        # ⚡ TRADUÇÃO DAS CHAVES: Converte a frase longa na sigla curta esperada pelo Go
        top_label_curto = LABEL_MAP[matched_desc_label]

        # Calcula o array de scores mockado para manter compatibilidade com o parser do Go
        scores = [1.0 if c == top_label_curto else 0.0 for c in chaves_curtas]

        logger.info(f"📤 SAÍDA   | Top Label Traduzida: '{top_label_curto}'")
        return ClassifyResp(labels=chaves_curtas, scores=scores, top_label=top_label_curto)

    except Exception as e:
        logger.error(f"❌ Erro durante a inferência GGUF: {e}")
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    import uvicorn
    # Mantém subindo na porta 8008 padrão do seu projeto
    uvicorn.run(app, host="0.0.0.0", port=8008)
