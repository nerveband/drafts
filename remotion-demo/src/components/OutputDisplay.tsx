import React from "react";

interface OutputDisplayProps {
  type: "json" | "list" | "create" | "tags";
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
      {type === "tags" && <TagsOutput />}
    </div>
  );
};

const JsonOutput: React.FC = () => (
  <pre style={{ margin: 0, fontFamily: "inherit", fontSize: 18, lineHeight: 1.5 }}>
    <span style={{ color: COLORS.bracket }}>{"{"}</span>
    {"\n  "}
    <span style={{ color: COLORS.key }}>"uuid"</span>
    <span>: </span>
    <span style={{ color: COLORS.string }}>"574FEA89..."</span>
    <span>,</span>
    {"\n  "}
    <span style={{ color: COLORS.key }}>"title"</span>
    <span>: </span>
    <span style={{ color: COLORS.string }}>"Meeting Notes"</span>
    <span>,</span>
    {"\n  "}
    <span style={{ color: COLORS.key }}>"tags"</span>
    <span>: </span>
    <span style={{ color: COLORS.bracket }}>{"["}</span>
    <span style={{ color: COLORS.string }}>"work"</span>
    <span>, </span>
    <span style={{ color: COLORS.string }}>"important"</span>
    <span style={{ color: COLORS.bracket }}>{"]"}</span>
    {"\n"}
    <span style={{ color: COLORS.bracket }}>{"}"}</span>
  </pre>
);

const ListOutput: React.FC = () => (
  <div style={{ fontFamily: "inherit", fontSize: 18 }}>
    <div style={{ color: COLORS.key, marginBottom: 12, fontWeight: 600 }}>
      {"UUID           TITLE                  FOLDER"}
    </div>
    <div style={{ borderTop: "1px solid #444", paddingTop: 12 }}>
      <div style={{ marginBottom: 8 }}>
        <span style={{ color: COLORS.string }}>574FEA89...  </span>
        <span>Weekly meeting notes   </span>
        <span style={{ color: COLORS.muted }}>inbox</span>
      </div>
      <div style={{ marginBottom: 8 }}>
        <span style={{ color: COLORS.string }}>A1B2C3D4...  </span>
        <span>Project roadmap        </span>
        <span style={{ color: COLORS.muted }}>inbox</span>
      </div>
      <div style={{ marginBottom: 8 }}>
        <span style={{ color: COLORS.string }}>E5F67890...  </span>
        <span>Shopping list          </span>
        <span style={{ color: COLORS.muted }}>inbox</span>
      </div>
    </div>
    <div style={{ marginTop: 12, color: COLORS.muted, fontSize: 16 }}>
      3 drafts found
    </div>
  </div>
);

const CreateOutput: React.FC = () => (
  <div style={{ fontFamily: "inherit", fontSize: 18, lineHeight: 1.5 }}>
    <div style={{ color: COLORS.success, fontSize: 20, fontWeight: 600, marginBottom: 12 }}>
      Draft created
    </div>
    <pre style={{ margin: 0, fontFamily: "inherit" }}>
      <span style={{ color: COLORS.bracket }}>{"{"}</span>
      {"\n  "}
      <span style={{ color: COLORS.key }}>"uuid"</span>
      <span>: </span>
      <span style={{ color: COLORS.string }}>"A1B2C3D4..."</span>
      <span>,</span>
      {"\n  "}
      <span style={{ color: COLORS.key }}>"title"</span>
      <span>: </span>
      <span style={{ color: COLORS.string }}>"New project idea"</span>
      <span>,</span>
      {"\n  "}
      <span style={{ color: COLORS.key }}>"tags"</span>
      <span>: </span>
      <span style={{ color: COLORS.bracket }}>{"["}</span>
      <span style={{ color: COLORS.string }}>"work"</span>
      <span style={{ color: COLORS.bracket }}>{"]"}</span>
      {"\n"}
      <span style={{ color: COLORS.bracket }}>{"}"}</span>
    </pre>
  </div>
);

const TagsOutput: React.FC = () => (
  <div style={{ fontFamily: "inherit", fontSize: 18 }}>
    <div style={{ marginBottom: 10 }}>
      <span style={{ color: COLORS.key, width: 120, display: "inline-block" }}>work</span>
      <span style={{ color: COLORS.muted }}>5 drafts</span>
    </div>
    <div style={{ marginBottom: 10 }}>
      <span style={{ color: COLORS.key, width: 120, display: "inline-block" }}>important</span>
      <span style={{ color: COLORS.muted }}>3 drafts</span>
    </div>
    <div style={{ marginBottom: 10 }}>
      <span style={{ color: COLORS.key, width: 120, display: "inline-block" }}>ideas</span>
      <span style={{ color: COLORS.muted }}>8 drafts</span>
    </div>
    <div style={{ marginTop: 16, color: COLORS.muted, fontSize: 16 }}>
      1 draft with tag "work"
    </div>
  </div>
);
