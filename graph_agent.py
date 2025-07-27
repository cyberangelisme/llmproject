from langgraph.graph import StateGraph, END
from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage
from langchain_core.prompts import ChatPromptTemplate
from typing import Dict, Any
from utils import *
from tool_node import *

# 设置 LLM
llm = get_chat_model()


# 创建图
workflow = StateGraph(GraphState)

# 添加节点
workflow.add_node("start", start_node)
workflow.add_node("decision", decision_node)
workflow.add_node("generate_func_node", generate_func_node)
workflow.add_node("retrival_func_node", retrival_func_node)
workflow.add_node("retrival_hyde_node", retrival_hyde_node)
workflow.add_node("ner_func_node", ner_func_node)
workflow.add_node("create_ferry_ticket_node",create_ferry_ticket_node)
workflow.add_node("mcp_amap_query", mcp_amap_query)
workflow.add_node("end", end_node)

# 设置入口点
workflow.set_entry_point("start")

# 添加边
workflow.add_edge("start", "decision")

# 添加条件边（由 LLM 决定）
workflow.add_conditional_edges(
    "decision",
    lambda x: x["next"],  # 从 state 中提取 "next" 键的值（即 Send 列表）
    []                    # 空列表表示不使用静态映射，允许动态 Send
    # #旧版本模式，预定义跳转
    # route_by_llm_decision,
    # {
    #     "generate_func_node": "generate_func_node",
    #     "retrival_func_node": "retrival_func_node",
    #     "ner_func_node": "ner_func_node",
    #     "retrival_hyde_node": "retrival_hyde_node",
    #     "create_ferry_ticket_node": "create_ferry_ticket_node",
    #     "mcp_amap_query": "mcp_amap_query",
    # }
)

# 固定边
workflow.add_edge("generate_func_node", "end")
workflow.add_edge("retrival_func_node", "end")
workflow.add_edge("ner_func_node", "end")
workflow.add_edge("retrival_hyde_node", "end")
workflow.add_edge("create_ferry_ticket_node", "end")
workflow.add_edge("mcp_amap_query","end")
# 编译图
app = workflow.compile()
app.get_graph().print_ascii()
# 测试运行
async def run_agent():
    inputs = {"input": "请创建一个medical的工单给ferry,并尝试查询当地有哪些医院"}  # 可以尝试 "hello world" 或 "hi there"
    result = await app.ainvoke(inputs)
    print("最终状态：", result)


if __name__ == "__main__":
    asyncio.run(run_agent())