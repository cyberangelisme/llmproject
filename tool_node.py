#from langgraph import AgentExecutor
from langchain.prompts import ChatPromptTemplate,PromptTemplate,MessagesPlaceholder
from langchain_core.messages import HumanMessage
from langchain.chains import LLMChain
from langchain_community.vectorstores.chroma import Chroma
from langchain.output_parsers import ResponseSchema,StructuredOutputParser
from langchain_google_community import GoogleSearchAPIWrapper
from langchain_mcp_adapters.client import MultiServerMCPClient
from langchain_core.tools import Tool
from langchain.chains.hyde.base import HypotheticalDocumentEmbedder
from langchain.agents import AgentExecutor,create_openai_functions_agent

import os
import asyncio
import time
from langgraph.types import Send
import httpx
from utils import *
from config import *
from prompt import *
from graph_state import GraphState
from model.ticket import *

AVAILABLE_TOOLS = [
    {"name": "generate_func_node", "description": "处理通用性质的问题，例如打招呼、闲聊、或者没有明确的工具可以处理的简单问题。"},
    {"name": "retrival_func_node", "description": "从内部文档库中检索信息，例如关于 'langchain是什么' 的问题。"},
    {"name": "retrival_hyde_node","description": "从内部知识库中检索信息，在用户提到hyde时使用。"},
    {"name": "ner_func_node", "description": "识别医疗实体，例如疾病、症状、药物（例如 '感冒吃什么好得快？阿莫西林可以吗？'）。"},
    #{"name": "custom_web_search", "description": "进行通用网页搜索，且不是医疗或地图相关（例如 'Python最新版本'）。"},
    #{"name": "Google Search", "description": "使用Google搜索，与 custom_web_search 类似，但可能更适合需要Google特定功能的情况。"},
    # {"name": "pharmacy_search", "description": "搜索附近的药店或医院（例如 'ip地址：111.40.58.236 附近有哪些 药店或者医院 ？'）。"},
    # {"name": "unsupported", "description": "如果问题无法被任何现有工具处理，请返回此项。"}
    {"name":"mcp_amap_query","description":"查询高德地图数据,有关地理信息的搜索"},
    {"name":"create_ferry_ticket_node","description":"创建任意主题的工单，构建post请求给ferry系统"}
]
# 动态生成工具列表字符串，用于注入到prompt中
TOOL_LIST_STR = "\n".join([f"- '{tool['name']}': {tool['description']}" for tool in AVAILABLE_TOOLS])

llm = get_chat_model()

# 节点函数
def start_node(state: GraphState) -> GraphState:
    print("Start node executed.")
    return {"input": state["input"], "output": "", "next_step": [""]}

def decision_node(state: GraphState) -> GraphState:
    print("Decision node executed. Letting LLM decide next step...")
    
    prompt = ChatPromptTemplate.from_messages([
        ("system", """你是一个流程决策助手。请根据用户的输入判断应该采取什么动作。
         格式要求为list[str], 比如：["工具1","使用工具2"]。没有则输出[],注意符号差异，必须使用 ""  而不是 ''
        {TOOL_LIST_STR}"""),
        ("user", "{input}")
    ])
    
    chain = prompt | llm
    response = chain.invoke({"input": state["input"],"TOOL_LIST_STR": TOOL_LIST_STR})
    print(response.content)
    next_nodes_list = json.loads(response.content.strip())
    # return {
    #     "input": state["input"],
    #     "output": "",
    #     "next_step": json.loads(response.content.strip())
    # }

    return {
        "next": [Send(node_name, state) for node_name in next_nodes_list]
    }
  


def calculator_node(state: GraphState) -> GraphState:
    print("Calculator node executed.")
    try:
        result = eval(state["input"])
    except Exception as e:
        result = f"计算错误: {e}"
    return {"input": state["input"], "output": str(result), "next_step": "end"}


def end_node(state: GraphState) -> GraphState:
    print("End node executed.")
    print("Final Output:", state["output"])
    return state


# 本地RAG检索召回
def retrival_func_node(state: GraphState)-> GraphState:
    query = state["input"]
    vdb = Chroma(
        embedding_function=get_embeddings_model(),
        persist_directory=os.path.join(os.path.dirname(__file__),'./data/db/')
    )
    documents = vdb.similarity_search_with_relevance_scores(query,k=5)
    # 保留概率大于0.7的查询结果
    query_result = [doc[0].page_content for doc in documents if doc[1]>0.7]
    prompt = PromptTemplate.from_template(RETRIEVAL_PROMPT_TPL)
    retrival_chain = LLMChain(
        llm=get_chat_model(),
        prompt=prompt,
        verbose = os.getenv("VERBOSE")
    )
    inputs = {
        'query': query,
        'query_result': '\n\n'.join(query_result) if len(query_result)>0 else "没有查询到结果"
    }
    result =  retrival_chain.run(inputs)
    return {
        'input': query,
        'output': result,
        'next_step': ['end'],
    }


# hyde 的rag 检索增强
def retrival_hyde_node(state: GraphState)-> GraphState:
    query = state['input']
    base_embeddings = get_embeddings_model()
    embeddings = HypotheticalDocumentEmbedder.from_llm(
        llm=get_chat_model(),
        base_embeddings=base_embeddings,
        prompt_key="web_search"
    )
    vdb = Chroma(
        embedding_function=embeddings,
        persist_directory=os.path.join(os.path.dirname(__file__),'./data/db/')
    )
    documents = vdb.similarity_search_with_relevance_scores(query,k=5)
    # 保留概率大于0.7的查询结果
    query_result = [doc[0].page_content for doc in documents if doc[1]>0.7]
    prompt = PromptTemplate.from_template(RETRIEVAL_PROMPT_TPL)
    retrival_chain = LLMChain(
        llm=get_chat_model(),
        prompt=prompt,
        verbose = os.getenv("VERBOSE")
    )
    inputs = {
        'query': query,
        'query_result': '\n\n'.join(query_result) if len(query_result)>0 else "没有查询到结果"
    }
    result =  retrival_chain.run(inputs)
    return {
        'input': query,
        'output': result,
        'next_step': ['end'],
    }


def generate_func_node(state: GraphState)-> GraphState:
    query = state['input']
    prompt = PromptTemplate.from_template(GENERIC_PROMPT_TPL)
    llm = get_chat_model()
    llm_chain = prompt | llm    
    result = llm_chain.invoke(query).content
    
    return{
        'input': query,
        'output':  result,
        # 'next_step': ['end']
    }
    
# 进行医疗实体识别，用于到知识图谱获取内容
def ner_func_node(state: GraphState)-> GraphState:
    query = state["input"]
    # 实体格式抽取定义
    response_schema = [
        ResponseSchema(type = 'list',name = 'disease',description='疾病名称实体'),
        ResponseSchema(type = 'list',name = 'symptom',description='疾病症状实体'),
        ResponseSchema(type = 'list',name = 'drug',description='疾病药品实体')  
    ]
    #json2struct out_put_parser.parse(json_str)
    out_put_parser = StructuredOutputParser(name="yiliao",response_schemas=response_schema)
    format_instructions = structed_output_parser(response_schema)


    prompt = PromptTemplate(
        template=NER_PROMPT_TPL,
        partial_variables={'format_instructions':format_instructions},
        input_variables=["query"], #动态传入参数
    )
    ner_chain = LLMChain(
        llm=get_chat_model(),
        prompt=prompt,
        verbose = os.getenv("VERBOSE")
    )
    result =  ner_chain.run(query)
    print(result)

    # jsonstr 2 dict
    ner_result = out_put_parser.parse(result)
    print(ner_result)
    return {
        "input": query,
        "output": ner_result,
        "next_step": "end"
    }

# 待修改，把langchain的agent形式改为自定义高德agent形式
#     # 使用官方集成的google tool,注意proxy_on
#     def search_google_func(self,query):
#         search = GoogleSearchAPIWrapper()
#         tool = Tool(
#             name="google_search",
#             description="Search Google for recent results.",
#             func=search.run,
#         )
#         query_result =  tool.run(query)
#         prompt = PromptTemplate(
#             template=SEARCH_PROMPT_TPL,
#             input_variables=["query"],
#             partial_variables={'query_result':query_result}
#         )
#         llm_chain = LLMChain(
#             llm=get_chat_model(),
#             prompt = prompt,
#             verbose=os.getenv("VERBOSE")
#         )
#         return llm_chain.run(query)


#     # 利用高德的mcp工具进行 附近药房搜索
#     def pharmacy_search_function(self,query):
#         #创建mcp client
            
#         result = asyncio.run(run_agent(query))
#         time.sleep(10)
#         return result

# 异步mcp函数定义，高德查询
async def mcp_amap_query(state : GraphState)->GraphState:
    query = state["input"]
    gaode_map_key = os.getenv("GAODE_MAP_KEY")
    print("gaode_map_key",gaode_map_key)
    llm_ds  = get_chat_model()
    client = MultiServerMCPClient(
        {
            "search": {
                "url": f"https://mcp.amap.com/sse?key={gaode_map_key}",
                "transport": "sse",
            }
        }
    ) 

    prompt = ChatPromptTemplate.from_messages([
        ("system", "你是一个有帮助的智能助手。请仔细分析用户问题并使用提供的工具来回答问题。"),
        ("user", "{input}"),
        MessagesPlaceholder(variable_name="agent_scratchpad"),
    ])
    print(prompt)
    # 获取工具并修剪描述长度
    tools = await client.get_tools()

    for tool in tools:
        print(f"工具名称：{tool.name}\n工具描述：{tool.description}\n 工具参数：{tool.args_schema}\n ")
    
    # 创建OpenAI函数代理
    agent = create_openai_functions_agent(
        llm=llm_ds,
        tools=tools,
        prompt=prompt,
    )
    # 创建代理执行器
    agent_executor = AgentExecutor(
        agent=agent,
        tools=tools,
        verbose=True,
        max_iterations=5,
        return_intermediate_steps=True,  # 返回中间步骤以便调试
        handle_parsing_errors=True,
    )
    
    # 运行代理
    agent_response = await agent_executor.ainvoke({
        "input": query
    })
    # 返回格式化的响应
    return {
        "input": state["input"],
        "output": str(agent_response),
        # "next_step": ["end"]
    }
    


#  构造工单函数，当LLM选择调用时，构建一个POST请求指向指定服务
async def create_ferry_ticket_node(state: GraphState) -> GraphState:
    print("create_ferry_ticket!")
    ferry_api_url = "http://localhost:8000/api/ferry"
    ferry_api_token = "my_secret_token"

    headers ={
        "Authorization": f"Bearer {ferry_api_token}",
        "Content-Type": "application/json"
    }

    try:
        # 创建工单数据
        ferry_ticket_request = FerryTicketData(
            title = "工单标题",
            content = state["input"],
            creator = "LangChain creator",
            priority = 1,
            category = "Medical"
        )
    except Exception as e:
        return "Error: " + str(e)
    
    # 发送工单数据
    async with httpx.AsyncClient() as client:
        response = await client.post(
            url=ferry_api_url,
            headers=headers,
            json=ferry_ticket_request.model_dump(),
        )
    print(response.status_code)
   

    return {
        "input": state["input"],
        "output": "",
        # "next_step": ["end"]
    }


    
# 条件边函数：由 LLM 的输出决定下一步
def route_by_llm_decision(state: GraphState) -> str:
    print(f"Routing to: {state['next_step']}")
    return state["next_step"]



