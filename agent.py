from utils import *
from config import *
from data_process import *  
from prompt import *
import os 
import asyncio
import time
from langchain_community.chains.llm_requests import LLMRequestsChain
from langchain.chains.llm import LLMChain
from langchain.prompts import PromptTemplate,ChatPromptTemplate,MessagesPlaceholder
from langchain.output_parsers import ResponseSchema,StructuredOutputParser
from langchain_core.tools import Tool
from langchain_google_community import GoogleSearchAPIWrapper
from langchain_mcp_adapters.client import MultiServerMCPClient
from langchain.agents import AgentExecutor,create_openai_functions_agent


# 异步mcp函数定义，高德查询
async def run_agent(query):
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
    #print(tools)
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
        "status": "success",
        "result": agent_response.get("output", ""),
        "steps": len(agent_response.get("intermediate_steps", [])),
    }
    
class Agent():
    def __init__(self):
        self.vdb = Chroma(
            embedding_function=get_embeddings_model(),
            persist_directory=os.path.join(os.path.dirname(__file__),'./data/db/')
        )


    # 对query定义一些预定义prompt，以及利用大模型的广泛生成性能力
    def generate_func(self,query):
        prompt = PromptTemplate.from_template(GENERIC_PROMPT_TPL)
        llm_chain = LLMChain(
            llm=get_chat_model(),
            prompt=prompt
        )
        return llm_chain.run(query)
    
    # RAG 利用文档来增强检索
    def retrival_func(self,query):
        documents = self.vdb.similarity_search_with_relevance_scores(query,k=5)
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
        return retrival_chain.run(inputs)
    
    # 进行医疗实体识别，用于到知识图谱获取内容
    def ner_func(self,query):
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

    # 自定义search tool
    def search_func(self, query: str):
            print(f"Executing search_func for query: '{query}'")
            # Construct the URL
            url = 'https://www.google.com/search?q=' + query.replace(' ', '+')
            
            # Fetch content using the utility function
            query_result = fetch_Google_Search_results_api(url)

            prompt = PromptTemplate(
                template=SEARCH_PROMPT_TPL,
                input_variables=["query", "query_result"], # Make sure this matches your template
            )
            llm_chain = LLMChain(
                llm=get_chat_model(), 
                prompt=prompt,
                verbose=os.getenv("VERBOSE") # Convert to boolean
            )
            inputs = {
                "query": query,
                "query_result": query_result # Pass the fetched content
            }
            return llm_chain.invoke(inputs)['text'] # Use invoke and get 'text'
    
    # 使用官方集成的google tool,注意proxy_on
    def search_google_func(self,query):
        search = GoogleSearchAPIWrapper()
        tool = Tool(
            name="google_search",
            description="Search Google for recent results.",
            func=search.run,
        )
        query_result =  tool.run(query)
        prompt = PromptTemplate(
            template=SEARCH_PROMPT_TPL,
            input_variables=["query"],
            partial_variables={'query_result':query_result}
        )
        llm_chain = LLMChain(
            llm=get_chat_model(),
            prompt = prompt,
            verbose=os.getenv("VERBOSE")
        )
        return llm_chain.run(query)


    # 利用高德的mcp工具进行 附近药房搜索
    def pharmacy_search_function(self,query):
        #创建mcp client
            
        result = asyncio.run(run_agent(query))
        time.sleep(10)
        return result

        # 执行入口，进行Agent创建和bind tools 工具绑定
    def query(self,query):
        tools =[
            Tool.from_function(
                name = 'generic_func',
                func= self.generic_func,
                description = '通用函数'
            )

        ]
if __name__ == "__main__":
    agent = Agent()
    
    #print(agent.generate_func("你是谁？"))
    #print(agent.retrival_func(query="langchain是什么？"))
    #print(agent.ner_func("感冒吃什么好得快？阿莫西林可以吗？"))
    #print(agent.search_google_func(query=" langchain是什么？"))
    print(agent.pharmacy_search_function("ip地址：111.40.58.236 附近有哪些 药店或者医院 ？"))
    # time.sleep(10)