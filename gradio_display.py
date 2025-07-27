#gradio use
import gradio as gr
import time
from agent import *
# def greet(name):
#     return "Hello " + name + "!"
# demo = gr.Interface(fn=greet, inputs="text", outputs="text")

# if __name__ == "__main__":
#     demo.launch(server_port=7862)



def process_input(user_text):
    """这个函数就是 Submit 按钮被点击后的回调逻辑。"""
    agent  = Agent()
    result = agent.pharmacy_search_function(user_text)
    # 对result 这个dict 进行解析
    #result_text = result['result'].replace('\n', '<br>')
    return result['result']

demo_interface = gr.Interface(
    fn=process_input,
    inputs=gr.Textbox(label="请输入文本"),
    outputs=gr.Textbox(label="处理结果"),
    title="Gradio Interface 默认 Submit 回调",
    description="点击 Submit 按钮，上方函数会被调用。"
)

if __name__ == "__main__":
    demo_interface.launch()