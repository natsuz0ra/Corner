/**
 * Spinner — Braille 字符旋转动画组件。
 * 使用 Braille 点阵图案序列创建旋转弧线效果。
 */

import React, { useEffect, useRef, useState } from "react";
import { Text } from "ink";

const FRAMES = ["⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"];
const FRAME_INTERVAL = 80; // 与 GradientFlowText 保持一致的刷新率

interface SpinnerProps {
  enabled: boolean;
  color?: string;
}

export function Spinner({ enabled, color = "#a78bfa" }: SpinnerProps): React.ReactElement {
  const [frameIndex, setFrameIndex] = useState(0);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    if (!enabled) {
      if (timerRef.current) {
        clearInterval(timerRef.current);
        timerRef.current = null;
      }
      return;
    }

    timerRef.current = setInterval(() => {
      setFrameIndex((prev) => (prev + 1) % FRAMES.length);
    }, FRAME_INTERVAL);

    return () => {
      if (timerRef.current) {
        clearInterval(timerRef.current);
        timerRef.current = null;
      }
    };
  }, [enabled]);

  if (!enabled) {
    return <Text />;
  }

  return (
    <Text bold color={color}>
      {FRAMES[frameIndex]}
    </Text>
  );
}
