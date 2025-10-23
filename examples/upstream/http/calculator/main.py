from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI()

class Numbers(BaseModel):
    a: int
    b: int

@app.post("/add")
async def add(numbers: Numbers):
    return {"result": numbers.a + numbers.b}

@app.post("/subtract")
async def subtract(numbers: Numbers):
    return {"result": numbers.a - numbers.b}

@app.post("/multiply")
async def multiply(numbers: Numbers):
    return {"result": numbers.a * numbers.b}

@app.post("/divide")
async def divide(numbers: Numbers):
    if numbers.b == 0:
        return {"error": "Cannot divide by zero"}
    return {"result": numbers.a / numbers.b}
