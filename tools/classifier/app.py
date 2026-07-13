import logging
import os
from typing import List
from fastapi import FastAPI, HTTPException
from huggingface_hub import hf_hub_download
from llama_cpp import Llama
from pydantic import BaseModel

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s | %(levelname)-8s | %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S"
)
logger = logging.getLogger("classifier")

app = FastAPI(title="Classifier Dynamic Router GGUF")

class ClassifyReq(BaseModel):
    text: str
    candidate_labels: List[str] # Agora este campo é obrigatório e dinâmico

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
            model_path = hf_hub_download(
                repo_id="Qwen/Qwen2.5-1.5B-Instruct-GGUF",
                filename="qwen2.5-1.5b-instruct-q4_k_m.gguf"
            )
            logger.info(f"📦 Carregando GGUF na memória: {model_path}")
            _llm = Llama(
                model_path=model_path,
                n_ctx=1024,
                n_threads=4,
                verbose=False
            )
            logger.info("✅ GGUF carregado e pronto para inferência!")
        except Exception as e:
            logger.error(f"❌ Falha ao inicializar o motor GGUF: {e}")
            raise RuntimeError(f"failed to initialize GGUF engine: {e}")
    return _llm

@app.post("/classify", response_model=ClassifyResp)
async def classify(req: ClassifyReq):
    if not req.text or not req.candidate_labels:
        raise HTTPException(status_code=400, detail="text and candidate_labels required")

    texto_limpo = req.text.strip().lower()

    # ⚡ FILTRO RÁPIDO: Se for health check, evita gastar CPU/GPU
    if "health check" in texto_limpo or texto_limpo == "ping":
        return ClassifyResp(labels=["GERAL"], scores=[1.0], top_label="GERAL")

    logger.info(f"📥 ENTRADA | Texto: '{req.text}'")

    # Formata as labels dinâmicas vindas do banco para o prompt do Qwen
    labels_str = "\n- ".join([f"'{l}'" for l in req.candidate_labels])

    prompt = f"""<|im_start|>system
Você é o Cérebro Roteador de um sistema multi-agentes. Sua função é analisar o input do usuário e decidir qual das categorias fornecidas melhor descreve o contexto da mensagem.

Diretrizes estritas:
1. Responda APENAS com o texto exato de uma das categorias fornecidas na lista abaixo. Nunca adicione explicações, pontuação ou markdown.
2. Escolha a categoria cujas palavras-chave e descrição melhor se alinhem com a intenção do usuário.

Categorias válidas disponíveis:
- {labels_str}
<|im_end|>
<|im_start|>user
Texto para analisar: "{req.text}"

Qual a categoria correta?<|im_end|>
<|im_start|>assistant
"""

    try:
        llm = get_llm()
        output = llm(prompt, max_tokens=64, temperature=0.0, stop=["<|im_end|>", "\n"])
        predicted_label = output["choices"][0]["text"].strip().replace("'", "").replace('"', '')
        logger.info(f"🤖 Resposta bruta do modelo: '{predicted_label}'")

        # Faz o match com a descrição enviada
        matched_label = req.candidate_labels[-1] # Fallback padrão inicial
        for label in req.candidate_labels:
            if label.lower() in predicted_label.lower() or predicted_label.lower() in label.lower():
                matched_label = label
                break

        # ⚡ EXTRAÇÃO AUTOMÁTICA DA CHAVE:
        # Se a label contiver ":", extrai apenas a chave curta (ex: "GO" ou "REACT")
        top_label_final = matched_label.split(":")[0].strip() if ":" in matched_label else matched_label

        # Limpa o array de labels de retorno para o Go receber apenas os IDs curtos
        chaves_curtas = [l.split(":")[0].strip() if ":" in l else l for l in req.candidate_labels]
        scores = [1.0 if c == top_label_final else 0.0 for c in chaves_curtas]

        logger.info(f"📤 SAÍDA   | Chave do Banco Selecionada: '{top_label_final}'")
        return ClassifyResp(labels=chaves_curtas, scores=scores, top_label=top_label_final)

    except Exception as e:
        logger.error(f"❌ Erro durante a inferência GGUF: {e}")
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    import uvicorn
    # Mantém subindo na porta 8008 padrão do seu projeto
    uvicorn.run(app, host="0.0.0.0", port=8008)
