from langchain_openai import OpenAIEmbeddings
from langchain_community.chat_models import ChatOpenAI
from langchain_google_genai import ChatGoogleGenerativeAI
from py2neo import Graph
from dotenv import load_dotenv
from config import *

import os
import requests
import json
load_dotenv()


def get_embeddings_model():
    model_map = {
        'openai':  OpenAIEmbeddings(
            model = os.getenv("OPENAI_EMBEDDING_MODEL"),
            base_url = os.getenv("BASE_URL")
        )
    }
    return model_map[os.getenv("EMBEDDING_MODEL")]

def get_chat_model():
    model_map = {
        'openai':  ChatOpenAI(
            model_name = os.getenv("OPENAI_CHAT_MODEL"),
            temperature = os.getenv("TEMPERATURE"),
            base_url = os.getenv("BASE_URL"),
            max_tokens = os.getenv("MAX_TOKENS")
        ),
        # 我不理解为什么gemini的模型在agent中会卡住
        'gemini': ChatGoogleGenerativeAI(
            #transport = "rest",
            model = os.getenv("GEMINI_CHAT_MODEL"),
            temperature = os.getenv("TEMPERATURE"),
            client_options={"api_endpoint": "https://api.openai-proxy.org/google"},
           # max_tokens = os.getenv("MAX_TOKENS"),
            google_api_key = os.getenv("GEMINI_API_KEY")
        )
    }
    return model_map[os.getenv("LLM_MODEL")]

#  实体结构化输出json
def structed_output_parser(response_schema):
    text = """
        请从以下文本中抽出实体类型信息，并确保json格式返回
        以下是字段含义和类型:

    """
    for schema in response_schema:
        text += schema.name + '字段表示：'+schema.description + '类型：'+schema.type+'\n'
    return text

# 替换模版占位符
def replace_template_placeholders(template, slots):
    for key, value in slots.items():
        template = template.replace('%'+key+'%', value)
    return template


# 未投入使用
def fetch_Google_Search_results_api(query: str, api_key: str, cx_id: str) -> str:
    """
    Fetches search results from Google Custom Search JSON API.
    Returns a summarized string of results, or an error message.
    """
    search_url = f"https://www.googleapis.com/customsearch/v1?key={api_key}&cx={cx_id}&q={query}"
    try:
        response = requests.get(search_url, timeout=10)
        response.raise_for_status()
        data = response.json()

        if "items" not in data:
            return "No relevant search results found via API."

        # 提取标题和片段作为 LLM 的输入
        results_summary = []
        for item in data["items"]:
            title = item.get("title", "No Title")
            snippet = item.get("snippet", "No Snippet")
            link = item.get("link", "#")
            results_summary.append(f"Title: {title}\nLink: {link}\nSnippet: {snippet}\n---")
        
        return "\n\n".join(results_summary[:5]) # 限制返回前5条结果

    except requests.exceptions.RequestException as e:
        print(f"Error calling Google Search API: {e}")
        return f"Error: Could not retrieve search results from API."
    except json.JSONDecodeError:
        print("Error: Could not decode JSON from Google Search API response.")
        return "Error: Invalid response from Google Search API."
    

if __name__ == "__main__":
    llm_model = get_chat_model()
    llm_embedding = get_embeddings_model()
    print(llm_model.invoke("你好"))
    print(llm_embedding.embed_query("你好"))