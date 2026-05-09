export type ApprovalMode = "standard" | "auto_review" | "auto";
export type ThinkingLevel = "off" | "low" | "medium" | "high" | "max";

export interface QAQuestion {
  id: string;
  question: string;
  options: string[];
  option_descriptions?: string[];
}

export interface QAAnswer {
  questionId: string;
  selectedOption: number;
  customAnswer: string;
}
