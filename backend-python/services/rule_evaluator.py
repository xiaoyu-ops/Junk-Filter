"""
RuleBasedEvaluator - 轻量级规则评估器

用途：
- LLM 不可用时的自动降级
- 快速评估（<10ms）
- 基于内容特征和关键词

特点：
- 关键词权重库（可配置）
- 内容长度奖励
- 结构化输出
"""

import logging
from typing import List, Dict, Any
from models.evaluation import EvaluationResult

logger = logging.getLogger(__name__)


class RuleBasedEvaluator:
    """基于关键词和规则的轻量级评估器"""

    def __init__(self):
        """初始化规则库"""

        # 创新度关键词（权重：高-3，中-2，低-1）
        self.innovation_keywords = {
            # 高创新 (3 分)
            'breakthrough': 3, 'revolutionary': 3, 'novel': 3,
            'first-time': 3, 'never-before': 3, 'paradigm shift': 3,
            'game-changing': 3, 'disruptive': 3, 'transformation': 3,

            # 中创新 (2 分)
            'innovation': 2, 'new approach': 2, 'discover': 2,
            'invented': 2, 'pioneering': 2, 'cutting-edge': 2,
            'next-generation': 2, 'advanced': 2, 'improved': 2,
            'novel approach': 2, 'innovation in': 2,

            # 低创新 (1 分)
            'research': 1, 'study': 1, 'analysis': 1, 'framework': 1,
            'algorithm': 1, 'method': 1, 'technique': 1, 'approach': 1,
        }

        # 深度关键词（权重：高-3，中-2，低-1）
        self.depth_keywords = {
            # 高深度 (3 分)
            'whitepaper': 3, 'comprehensive analysis': 3, 'in-depth': 3,
            'deep dive': 3, 'peer-reviewed': 3, 'academic': 3,
            'research paper': 3, 'rigorous': 3, 'systematic review': 3,

            # 中深度 (2 分)
            'detailed': 2, 'thorough': 2, 'investigation': 2,
            'case study': 2, 'analysis': 2, 'research': 2,
            'technical': 2, 'scientific': 2, 'evidence': 2,

            # 低深度 (1 分)
            'overview': 1, 'summary': 1, 'introduction': 1,
            'guide': 1, 'tutorial': 1, 'brief': 1, 'quick': 1,
        }

        # 领域权重（不同领域的重要性）
        self.domain_weights = {
            'ai': 1.2, 'machine learning': 1.2, 'deep learning': 1.2,
            'nlp': 1.1, 'computer vision': 1.1, 'neural': 1.1,
            'blockchain': 1.0, 'cryptocurrency': 1.0, 'web3': 1.0,
            'cloud': 0.9, 'kubernetes': 0.9, 'devops': 0.9,
            'security': 1.0, 'cryptography': 1.0, 'encryption': 1.0,
            'performance': 0.8, 'optimization': 0.8, 'benchmarks': 0.8,
        }

    def evaluate(
        self,
        title: str,
        content: str,
        url: str = ""
    ) -> EvaluationResult:
        """
        基于规则的评估

        返回：EvaluationResult
        """
        # 预处理文本
        text = f"{title} {content}".lower()

        # 计算创新度评分 (0-10)
        innovation_score = self._calculate_score(
            text,
            self.innovation_keywords,
            base=3.0  # 基础分
        )

        # 计算深度评分 (0-10)
        depth_score = self._calculate_score(
            text,
            self.depth_keywords,
            base=3.0
        )

        # 内容长度奖励
        content_length_bonus = self._calculate_length_bonus(content)
        depth_score = min(10, depth_score + content_length_bonus)

        # 标题长度奖励（更长的标题通常更信息丰富）
        title_length_bonus = min(1.0, len(title.split()) / 15)
        innovation_score = min(10, innovation_score + title_length_bonus)

        # 领域权重调整
        domain_weight = self._get_domain_weight(text)
        innovation_score = min(10, innovation_score * domain_weight)

        # 决策逻辑
        decision = self._make_decision(innovation_score, depth_score)

        # 提取关键概念
        key_concepts = self._extract_concepts(text)

        # 生成 TLDR
        tldr = self._generate_tldr(title, content, len(content.split()))

        # 生成推理说明
        reasoning = self._generate_reasoning(innovation_score, depth_score, decision)

        logger.debug(
            f"[Rule Eval] "
            f"Innovation: {innovation_score:.1f}, "
            f"Depth: {depth_score:.1f}, "
            f"Decision: {decision}"
        )

        return EvaluationResult(
            innovation_score=int(min(10, max(0, innovation_score))),
            depth_score=int(min(10, max(0, depth_score))),
            decision=decision,
            reasoning=reasoning,
            tldr=tldr,
            key_concepts=key_concepts
        )

    def _calculate_score(
        self,
        text: str,
        keyword_dict: Dict[str, int],
        base: float = 0.0
    ) -> float:
        """计算基于关键词的评分"""
        score = base
        matched_keywords = []

        for keyword, weight in keyword_dict.items():
            if keyword in text:
                score += weight
                matched_keywords.append((keyword, weight))

        # 对数缩放，防止单个关键词影响过大
        # 基于 importance，权重越高的关键词权重衰减越慢
        for keyword, weight in matched_keywords[:5]:  # 只计算前 5 个最相关的
            if weight <= 1:
                score += weight * 0.5  # 低权重关键词，贡献减半
            elif weight == 2:
                score += weight * 0.7  # 中权重关键词，贡献 70%
            # 高权重关键词全额计入

        return min(10, score)

    def _calculate_length_bonus(self, content: str) -> float:
        """根据内容长度计算深度奖励"""
        word_count = len(content.split())

        if word_count < 100:
            return 0.0  # 太短
        elif word_count < 300:
            return 0.5
        elif word_count < 800:
            return 1.0
        elif word_count < 1500:
            return 1.5
        else:
            return 2.0  # 长文章通常深度更大

    def _get_domain_weight(self, text: str) -> float:
        """获取领域权重（AI/ML 内容权重更高）"""
        for domain, weight in self.domain_weights.items():
            if domain in text:
                return weight
        return 1.0  # 默认权重

    def _make_decision(self, innovation: float, depth: float) -> str:
        """基于评分制定决策"""
        # 决策树
        if innovation >= 7 and depth >= 6:
            return "INTERESTING"
        elif innovation >= 6 and depth >= 5:
            return "INTERESTING"
        elif innovation >= 5 or depth >= 6:
            return "BOOKMARK"
        elif innovation >= 3 or depth >= 3:
            return "BOOKMARK"
        else:
            return "SKIP"

    def _extract_concepts(self, text: str) -> List[str]:
        """提取关键概念"""
        concepts = []

        # 预定义概念检测
        concept_indicators = {
            'AI/ML': ['ai', 'artificial intelligence', 'machine learning', 'deep learning', 'neural'],
            'Data': ['data', 'database', 'analytics', 'big data', 'dataset'],
            'Cloud': ['cloud', 'aws', 'azure', 'kubernetes', 'docker'],
            'Security': ['security', 'encryption', 'cryptography', 'auth', 'vulnerability'],
            'Web': ['web', 'frontend', 'backend', 'api', 'rest', 'graphql'],
            'DevOps': ['devops', 'cicd', 'deployment', 'infrastructure', 'monitoring'],
            'Performance': ['performance', 'optimization', 'benchmark', 'latency', 'throughput'],
        }

        for concept, indicators in concept_indicators.items():
            for indicator in indicators:
                if indicator in text:
                    concepts.append(concept)
                    break

        # 返回前 5 个概念
        return concepts[:5]

    def _generate_tldr(self, title: str, content: str, word_count: int) -> str:
        """生成一句话总结"""
        # 简单的 TLDR 生成策略
        if word_count > 1000:
            return f"In-depth analysis: {title}"
        elif word_count > 500:
            return f"Detailed exploration: {title}"
        elif word_count > 200:
            return f"Discussion about {title}"
        else:
            return f"Note on {title}"

    def _generate_reasoning(
        self,
        innovation: float,
        depth: float,
        decision: str
    ) -> str:
        """生成推理说明"""
        parts = [
            f"Innovation score: {innovation:.1f}/10",
            f"Depth score: {depth:.1f}/10"
        ]

        if decision == "INTERESTING":
            parts.append("High innovation and depth - recommended for priority reading")
        elif decision == "BOOKMARK":
            parts.append("Moderate innovation or depth - worth bookmarking for later")
        else:
            parts.append("Low innovation and depth - may skip unless directly relevant")

        return " | ".join(parts)
