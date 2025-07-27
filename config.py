# config.py
GRAPH_TEMPLATE = {
    # 1. 疾病定义
    'desc': {
        'slots': ['disease'],
        'question': '什么叫%disease%？ / %disease%是一种什么病？',
        'cypher': "MATCH (n:Disease {name: '%disease%'}) RETURN n.desc AS RES",
        'answer': '【%disease%】的定义：%RES%',
    },
    
    # 2. 疾病病因
    'cause': {
        'slots': ['disease'],
        'question': '%disease%一般是由什么引起的？ / 什么会导致%disease%？',
        'cypher': "MATCH (n:Disease {name: '%disease%'})-[:CAUSED_BY]->(c) RETURN c.name AS RES",
        'answer': '【%disease%】的病因：%RES%',
    },
    
    # 3. 疾病症状
    'disease_symptom': {
        'slots': ['disease'],
        'question': '%disease%会有哪些症状？ / %disease%有哪些临床表现？',
        'cypher': "MATCH (n:Disease {name: '%disease%'})-[:HAS_SYMPTOM]->(s) RETURN s.name AS RES",
        'answer': '【%disease%】的症状：%RES%',
    },
    
    # 4. 疾病推荐药物
    'treatment_drug': {
        'slots': ['disease'],
        'question': '%disease%应该吃什么药？ / %disease%的常用药物有哪些？',
        'cypher': "MATCH (n:Disease {name: '%disease%'})-[:TREAT_WITH]->(d:Drug) RETURN d.name AS RES",
        'answer': '【%disease%】的推荐药物：%RES%',
    },
    
    # 5. 疾病所属科室
    'department': {
        'slots': ['disease'],
        'question': '%disease%应该挂哪个科？ / %disease%属于哪个科室？',
        'cypher': "MATCH (n:Disease {name: '%disease%'})-[:BELONG_TO]->(dept:Department) RETURN dept.name AS RES",
        'answer': '【%disease%】所属科室：%RES%',
    },
    
    # 6. 药物禁忌症
    'drug_contraindication': {
        'slots': ['drug'],
        'question': '%drug%不能和哪些药一起吃？ / %drug%的禁忌药有哪些？',
        'cypher': "MATCH (d:Drug {name: '%drug%'})-[:CONTRAINDICATED_WITH]->(c:Drug) RETURN c.name AS RES",
        'answer': '【%drug%】的禁忌药物：%RES%',
    },
    
    # 7. 药物适应症
    'drug_indication': {
        'slots': ['drug'],
        'question': '%drug%可以治疗什么病？ / %drug%适用于哪些疾病？',
        'cypher': "MATCH (d:Drug {name: '%drug%'})<-[:TREAT_WITH]-(n:Disease) RETURN n.name AS RES",
        'answer': '【%drug%】可以治疗的疾病：%RES%',
    },
    
    # 8. 综合信息查询
    'full_info': {
        'slots': ['disease'],
        'question': '请介绍一下%disease% / %disease%的相关信息？',
        'cypher': """
            MATCH (d:Disease {name: '%disease%'})
            OPTIONAL MATCH (d)-[:HAS_SYMPTOM]->(s:Symptom)
            OPTIONAL MATCH (d)-[:TREAT_WITH]->(dr:Drug)
            OPTIONAL MATCH (d)-[:BELONG_TO]->(dept:Department)
            RETURN 
                d.desc AS desc,
                collect(DISTINCT s.name) AS symptoms,
                collect(DISTINCT dr.name) AS drugs,
                collect(DISTINCT dept.name) AS departments
        """,
        'answer': """
            【%disease%】的定义：%desc%
            - 主要症状：%symptoms%
            - 推荐药物：%drugs%
            - 所属科室：%departments%
        """,
    },
}