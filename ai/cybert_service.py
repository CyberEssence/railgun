# -*- coding: utf-8 -*-
from fastapi import FastAPI
from pydantic import BaseModel
from transformers import pipeline, AutoModelForSequenceClassification, AutoTokenizer
import uvicorn

app = FastAPI()

print("Loading Cybert model...")
model_name = "SynamicTechnologies/CYBERT"
classifier = None
try:
    tokenizer = AutoTokenizer.from_pretrained(model_name)
    model = AutoModelForSequenceClassification.from_pretrained(model_name)
    classifier = pipeline("text-classification", model=model, tokenizer=tokenizer)
    print("Model loaded!")
except Exception as e:
    print("Error loading model: " + str(e))

class LogRequest(BaseModel):
    log_text: str

@app.post("/analyze")
async def analyze_log(req: LogRequest):
    if not classifier:
        return {"error": "Model not loaded"}
    
    result = classifier(req.log_text)[0]
    
    return {
        "label": result['label'],
        "score": float(result['score']),
        "is_malicious": "malicious" in result['label'].lower() or result['score'] > 0.9
    }

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8001)