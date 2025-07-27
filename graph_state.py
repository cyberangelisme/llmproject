from typing import Annotated,List,Dict
import operator

# 定义状态
class GraphState(Dict):
    origin_query: str
    input: Annotated[str, operator.add]
    output: Annotated[str, operator.add]
    #next_step: List[str]   # 实际不影响决策，仅仅是context上下文内容的保持，决策依赖于langgraph 
