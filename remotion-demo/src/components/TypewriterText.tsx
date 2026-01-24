import React from "react";
import { interpolate } from "remotion";

interface TypewriterTextProps {
  text: string;
  startFrame: number;
  duration: number;
  frameInSequence: number;
  color: string;
}

export const TypewriterText: React.FC<TypewriterTextProps> = ({
  text,
  startFrame,
  duration,
  frameInSequence,
  color,
}) => {
  const progress = interpolate(
    frameInSequence,
    [startFrame, startFrame + duration],
    [0, 1],
    { extrapolateLeft: "clamp", extrapolateRight: "clamp" }
  );

  const visibleChars = Math.floor(progress * text.length);
  const displayText = text.slice(0, visibleChars);

  return (
    <span style={{ color, fontWeight: 500 }}>
      {displayText}
    </span>
  );
};
