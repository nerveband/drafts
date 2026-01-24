import { Composition } from "remotion";
import { DraftsCliDemo } from "./DraftsCliDemo";

export const RemotionRoot: React.FC = () => {
  return (
    <>
      <Composition
        id="DraftsCliDemo"
        component={DraftsCliDemo}
        durationInFrames={240}
        fps={30}
        width={1280}
        height={720}
      />
    </>
  );
};
