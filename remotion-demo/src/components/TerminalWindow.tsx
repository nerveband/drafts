import React from "react";
import { FONT_FAMILY } from "../fonts";

// Beeper colors - flat
const COLORS = {
  terminal: "#1a1a2e",
  titleBar: "#252540",
};

interface TerminalWindowProps {
  children: React.ReactNode;
}

export const TerminalWindow: React.FC<TerminalWindowProps> = ({ children }) => {
  return (
    <div
      style={{
        borderRadius: 12,
        overflow: "hidden",
        background: COLORS.terminal,
        border: "2px solid #3a3a5a",
      }}
    >
      {/* Title bar - flat */}
      <div
        style={{
          height: 44,
          background: COLORS.titleBar,
          borderBottom: "1px solid #3a3a5a",
          display: "flex",
          alignItems: "center",
          padding: "0 16px",
          position: "relative",
        }}
      >
        {/* Traffic lights */}
        <div style={{ display: "flex", gap: 8, zIndex: 1 }}>
          <TrafficLight color="#ff5f57" />
          <TrafficLight color="#febc2e" />
          <TrafficLight color="#28c840" />
        </div>

        {/* Title */}
        <div
          style={{
            position: "absolute",
            left: "50%",
            transform: "translateX(-50%)",
            color: "#888",
            fontSize: 13,
            fontWeight: 500,
            fontFamily: FONT_FAMILY,
          }}
        >
          drafts - Terminal
        </div>
      </div>

      {/* Terminal content */}
      <div
        style={{
          padding: "28px 40px",
          fontSize: 24,
          lineHeight: 1.7,
          color: "#e8e8e8",
          minHeight: 440,
          fontFamily: FONT_FAMILY,
        }}
      >
        {children}
      </div>
    </div>
  );
};

interface TrafficLightProps {
  color: string;
}

const TrafficLight: React.FC<TrafficLightProps> = ({ color }) => (
  <div
    style={{
      width: 12,
      height: 12,
      borderRadius: "50%",
      backgroundColor: color,
    }}
  />
);
