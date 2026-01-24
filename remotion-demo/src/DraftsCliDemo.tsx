import React, { useEffect } from "react";
import {
  AbsoluteFill,
  useCurrentFrame,
  useVideoConfig,
  interpolate,
  delayRender,
  continueRender,
} from "remotion";
import { TerminalWindow } from "./components/TerminalWindow";
import { TypewriterText } from "./components/TypewriterText";
import { OutputDisplay } from "./components/OutputDisplay";
import { loadFonts, FONT_FAMILY } from "./fonts";

// Beeper color palette
const COLORS = {
  white: "#ffffff",
  primary: "#003471",
  secondary: "#6288b5",
  dark: "#173973",
  gradientStart: "#0f417e",
  gradientEnd: "#3d689b",
  command: "#6bcbff",
  accent: "#ff9f43",
  text: "#e8e8e8",
};

// Demo sequences - each shows a different feature
const SEQUENCES = [
  {
    command: "drafts list",
    output: "list",
    description: "List All Drafts",
    icon: "[]",
  },
  {
    command: "drafts new \"New project idea\" -t work",
    output: "create",
    description: "Create with Tags",
    icon: "+",
  },
  {
    command: "drafts get 574FEA89",
    output: "json",
    description: "JSON Output",
    icon: "{}",
  },
  {
    command: "drafts list -t work",
    output: "tags",
    description: "Filter by Tags",
    icon: "#",
  },
];

export const DraftsCliDemo: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps, durationInFrames } = useVideoConfig();

  // Load fonts
  const [handle] = React.useState(() => delayRender());
  useEffect(() => {
    loadFonts().then(() => continueRender(handle));
  }, [handle]);

  // Each sequence lasts about 2.5 seconds (75 frames at 30fps) - faster pacing
  const framesPerSequence = Math.floor(durationInFrames / SEQUENCES.length);
  const currentSequenceIndex = Math.floor(frame / framesPerSequence) % SEQUENCES.length;
  const frameInSequence = frame % framesPerSequence;

  const currentSequence = SEQUENCES[currentSequenceIndex];

  // Animation timing within each sequence - faster, no fades
  const typewriterDuration = 18; // faster typing
  const outputDelay = 22; // show output sooner

  // No fade transitions - clean cuts (always full opacity)
  const sequenceOpacity = 1;

  return (
    <AbsoluteFill
      style={{
        background: COLORS.primary,
        fontFamily: FONT_FAMILY,
      }}
    >

      {/* Title */}
      <div
        style={{
          position: "absolute",
          top: 40,
          left: 0,
          right: 0,
          textAlign: "center",
        }}
      >
        <h1
          style={{
            fontSize: 48,
            fontWeight: 700,
            color: COLORS.white,
            margin: 0,
          }}
        >
          Drafts CLI
        </h1>
        <p
          style={{
            fontSize: 18,
            color: COLORS.secondary,
            marginTop: 6,
          }}
        >
          Edit your Drafts from the command line
        </p>
      </div>

      {/* Terminal Window */}
      <div
        style={{
          position: "absolute",
          top: 150,
          left: "50%",
          transform: "translateX(-50%)",
          width: 1100,
        }}
      >
        <TerminalWindow>
          {/* Command Line */}
          <div style={{ marginBottom: 28, fontSize: 28 }}>
            <span style={{ color: COLORS.accent, marginRight: 14, fontWeight: 600 }}>$</span>
            <TypewriterText
              text={currentSequence.command}
              startFrame={3}
              duration={typewriterDuration}
              frameInSequence={frameInSequence}
              color={COLORS.command}
            />
            <Cursor
              visible={frameInSequence >= 3 && frameInSequence < typewriterDuration + 3}
              color={COLORS.command}
            />
          </div>

          {/* Output - instant appear, no fade */}
          {frameInSequence >= outputDelay && (
            <OutputDisplay
              type={currentSequence.output as "json" | "list" | "create" | "tags"}
              fadeInProgress={1}
            />
          )}
        </TerminalWindow>
      </div>

      {/* Feature indicator */}
      <div
        style={{
          position: "absolute",
          bottom: 16,
          left: "50%",
          transform: "translateX(-50%)",
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          gap: 10,
        }}
      >
        {/* Feature label */}
        <div
          style={{
            fontSize: 16,
            color: COLORS.white,
            fontWeight: 600,
            letterSpacing: "0.05em",
            textTransform: "uppercase",
          }}
        >
          {currentSequence.description}
        </div>

        {/* Progress dots - simple */}
        <div style={{ display: "flex", gap: 12 }}>
          {SEQUENCES.map((seq, index) => {
            const isActive = index === currentSequenceIndex;
            return (
              <div
                key={index}
                style={{
                  width: isActive ? 32 : 10,
                  height: 10,
                  borderRadius: 5,
                  backgroundColor: isActive ? COLORS.accent : COLORS.secondary,
                }}
              />
            );
          })}
        </div>
      </div>
    </AbsoluteFill>
  );
};

// Blinking cursor component
interface CursorProps {
  visible: boolean;
  color: string;
}

const Cursor: React.FC<CursorProps> = ({ visible, color }) => {
  const frame = useCurrentFrame();
  const blinkVisible = Math.floor(frame / 15) % 2 === 0;

  if (!visible) return null;

  return (
    <span
      style={{
        display: "inline-block",
        width: 3,
        height: 28,
        backgroundColor: color,
        marginLeft: 4,
        opacity: blinkVisible ? 1 : 0,
        verticalAlign: "middle",
      }}
    />
  );
};
