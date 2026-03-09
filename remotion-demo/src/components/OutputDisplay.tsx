import React from "react";

interface OutputDisplayProps {
  type: "json" | "list" | "create" | "actions";
  fadeInProgress: number;
}

// Simple flat colors
const COLORS = {
  string: "#7ec87e",
  key: "#6bb8ff",
  bracket: "#b794f6",
  success: "#4ade80",
  muted: "#888",
};

export const OutputDisplay: React.FC<OutputDisplayProps> = ({
  type,
  fadeInProgress,
}) => {
  return (
    <div
      style={{
        opacity: fadeInProgress,
        transform: `translateY(${(1 - fadeInProgress) * 10}px)`,
      }}
    >
      {type === "json" && <JsonOutput />}
      {type === "list" && <ListOutput />}
      {type === "create" && <CreateOutput />}
      {type === "actions" && <ActionsOutput />}
    </div>
  );
};

const JsonOutput: React.FC = () => (
  <pre style={{ margin: 0, fontFamily: "inherit", fontSize: 18, lineHeight: 1.5 }}>
    <span style={{ color: COLORS.bracket }}>{"{"}</span>
    {"\n  "}
    <span style={{ color: COLORS.key }}>"success"</span>
    <span>: </span>
    <span>true</span>
    <span>,</span>
    {"\n  "}
    <span style={{ color: COLORS.key }}>"data"</span>
    <span>: </span>
    <span style={{ color: COLORS.bracket }}>{"{"}</span>
    {"\n    "}
    <span style={{ color: COLORS.key }}>"uuid"</span>
    <span>: </span>
    <span style={{ color: COLORS.string }}>"574FEA89..."</span>
    <span>,</span>
    {"\n    "}
    <span style={{ color: COLORS.key }}>"title"</span>
    <span>: </span>
    <span style={{ color: COLORS.string }}>"Meeting Notes"</span>
    <span>,</span>
    {"\n    "}
    <span style={{ color: COLORS.key }}>"tags"</span>
    <span>: </span>
    <span style={{ color: COLORS.bracket }}>{"["}</span>
    <span style={{ color: COLORS.string }}>"work"</span>
    <span>, </span>
    <span style={{ color: COLORS.string }}>"important"</span>
    <span style={{ color: COLORS.bracket }}>{"]"}</span>
    {"\n  "}
    <span style={{ color: COLORS.bracket }}>{"}"}</span>
    {"\n"}
    <span style={{ color: COLORS.bracket }}>{"}"}</span>
  </pre>
);

const ListOutput: React.FC = () => (
  <pre style={{ margin: 0, fontFamily: "inherit", fontSize: 18, lineHeight: 1.5 }}>
    <span style={{ color: COLORS.bracket }}>{"{"}</span>
    {"\n  "}
    <span style={{ color: COLORS.key }}>"success"</span>
    <span>: </span>
    <span>true</span>
    <span>,</span>
    {"\n  "}
    <span style={{ color: COLORS.key }}>"data"</span>
    <span>: </span>
    <span style={{ color: COLORS.bracket }}>{"{"}</span>
    {"\n    "}
    <span style={{ color: COLORS.key }}>"drafts"</span>
    <span>: </span>
    <span style={{ color: COLORS.bracket }}>{"["}</span>
    {"\n      "}
    <span style={{ color: COLORS.bracket }}>{"{"}</span>
    <span style={{ color: COLORS.key }}>"uuid"</span>
    <span>: </span>
    <span style={{ color: COLORS.string }}>"574FEA89..."</span>
    <span>, </span>
    <span style={{ color: COLORS.key }}>"title"</span>
    <span>: </span>
    <span style={{ color: COLORS.string }}>"Weekly meeting notes"</span>
    <span>, </span>
    <span style={{ color: COLORS.key }}>"folder"</span>
    <span>: </span>
    <span style={{ color: COLORS.string }}>"inbox"</span>
    <span style={{ color: COLORS.bracket }}>{"}"}</span>
    <span>,</span>
    {"\n      "}
    <span style={{ color: COLORS.bracket }}>{"{"}</span>
    <span style={{ color: COLORS.key }}>"uuid"</span>
    <span>: </span>
    <span style={{ color: COLORS.string }}>"A1B2C3D4..."</span>
    <span>, </span>
    <span style={{ color: COLORS.key }}>"title"</span>
    <span>: </span>
    <span style={{ color: COLORS.string }}>"Project roadmap"</span>
    <span>, </span>
    <span style={{ color: COLORS.key }}>"folder"</span>
    <span>: </span>
    <span style={{ color: COLORS.string }}>"inbox"</span>
    <span style={{ color: COLORS.bracket }}>{"}"}</span>
    {"\n    "}
    <span style={{ color: COLORS.bracket }}>{"]"}</span>
    <span>,</span>
    {"\n    "}
    <span style={{ color: COLORS.key }}>"count"</span>
    <span>: </span>
    <span>2</span>
    <span>,</span>
    {"\n    "}
    <span style={{ color: COLORS.key }}>"limit"</span>
    <span>: </span>
    <span>3</span>
    <span>,</span>
    {"\n    "}
    <span style={{ color: COLORS.key }}>"full"</span>
    <span>: </span>
    <span>false</span>
    {"\n  "}
    <span style={{ color: COLORS.bracket }}>{"}"}</span>
    {"\n"}
    <span style={{ color: COLORS.bracket }}>{"}"}</span>
  </pre>
);

const CreateOutput: React.FC = () => (
  <pre style={{ margin: 0, fontFamily: "inherit", fontSize: 18, lineHeight: 1.5 }}>
    <span style={{ color: COLORS.bracket }}>{"{"}</span>
    {"\n  "}
    <span style={{ color: COLORS.key }}>"success"</span>
    <span>: </span>
    <span>true</span>
    <span>,</span>
    {"\n  "}
    <span style={{ color: COLORS.key }}>"data"</span>
    <span>: </span>
    <span style={{ color: COLORS.bracket }}>{"{"}</span>
    {"\n    "}
    <span style={{ color: COLORS.key }}>"uuid"</span>
    <span>: </span>
    <span style={{ color: COLORS.string }}>"A1B2C3D4..."</span>
    <span>,</span>
    {"\n    "}
    <span style={{ color: COLORS.key }}>"title"</span>
    <span>: </span>
    <span style={{ color: COLORS.string }}>"New project idea"</span>
    <span>,</span>
    {"\n    "}
    <span style={{ color: COLORS.key }}>"tags"</span>
    <span>: </span>
    <span style={{ color: COLORS.bracket }}>{"["}</span>
    <span style={{ color: COLORS.string }}>"work"</span>
    <span style={{ color: COLORS.bracket }}>{"]"}</span>
    <span>,</span>
    {"\n    "}
    <span style={{ color: COLORS.key }}>"content"</span>
    <span>: </span>
    <span style={{ color: COLORS.string }}>"New project idea"</span>
    {"\n  "}
    <span style={{ color: COLORS.bracket }}>{"}"}</span>
    {"\n"}
    <span style={{ color: COLORS.bracket }}>{"}"}</span>
  </pre>
);

const ActionsOutput: React.FC = () => (
  <pre style={{ margin: 0, fontFamily: "inherit", fontSize: 18, lineHeight: 1.5 }}>
    <span style={{ color: COLORS.bracket }}>{"{"}</span>
    {"\n  "}
    <span style={{ color: COLORS.key }}>"success"</span>
    <span>: </span>
    <span>true</span>
    <span>,</span>
    {"\n  "}
    <span style={{ color: COLORS.key }}>"data"</span>
    <span>: </span>
    <span style={{ color: COLORS.bracket }}>{"{"}</span>
    {"\n    "}
    <span style={{ color: COLORS.key }}>"actions"</span>
    <span>: </span>
    <span style={{ color: COLORS.bracket }}>{"["}</span>
    <span style={{ color: COLORS.string }}>"Copy"</span>
    <span>, </span>
    <span style={{ color: COLORS.string }}>"Copy Text"</span>
    <span>, </span>
    <span style={{ color: COLORS.string }}>"Copy as HTML"</span>
    <span style={{ color: COLORS.bracket }}>{"]"}</span>
    <span>,</span>
    {"\n    "}
    <span style={{ color: COLORS.key }}>"count"</span>
    <span>: </span>
    <span>3</span>
    <span>,</span>
    {"\n    "}
    <span style={{ color: COLORS.key }}>"search"</span>
    <span>: </span>
    <span style={{ color: COLORS.string }}>"Copy"</span>
    {"\n  "}
    <span style={{ color: COLORS.bracket }}>{"}"}</span>
    {"\n"}
    <span style={{ color: COLORS.bracket }}>{"}"}</span>
  </pre>
);
