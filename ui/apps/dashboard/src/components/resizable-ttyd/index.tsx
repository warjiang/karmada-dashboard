import { Icons } from '@/components/icons';
import { Direction } from 're-resizable/lib/resizer';
import { NumberSize, Resizable } from 're-resizable';
import {
  ContainerTerminal,
  FitAddon,
  TerminalContext,
} from '@karmada/terminal';
import { useWindowSize } from '@uidotdev/usehooks';
import { useContext, useEffect, useRef } from 'react';
import { useAuth } from '@/components/auth';

function TopHandle() {
  return (
    <div
      id="top-handle"
      className="w-full h-[5px] flex justify-center items-center bg-gray-400"
    >
      <Icons.more height={20} />
    </div>
  );
}

const ResizableTtyd = () => {
  const { authenticated } = useAuth();
  const size = useWindowSize();
  const terminalContextData = useContext(TerminalContext);
  const terminal = terminalContextData.terminal as ContainerTerminal;
  const { showTerminal } = terminalContextData;
  const terminalContainerRef = useRef<HTMLDivElement | null>(null);
  useEffect(() => {
    if (!terminalContainerRef.current || !authenticated || !showTerminal)
      return;
    console.log('init terminal');
    // eslint-disable-next-line @typescript-eslint/no-floating-promises
    terminal.getSessionId().then(() => {
      terminal.open(terminalContainerRef.current!);
      terminal.connect();
    });
  }, [terminalContainerRef.current, authenticated, showTerminal]);
  return (
    <Resizable
      handleComponent={{
        top: <TopHandle />,
      }}
      defaultSize={{
        height: 'auto',
      }}
      size={{
        width: size.width || 1024,
      }}
      className={'w-full'}
      style={{
        position: 'absolute',
        bottom: 0,
        visibility: showTerminal ? 'visible' : 'hidden',
      }}
      onResizeStop={(
        event: MouseEvent | TouchEvent,
        direction: Direction,
        elementRef: HTMLElement,
        delta: NumberSize,
      ) => {
        console.log('event', event);
        console.log('direction', direction);
        console.log('elementRef', elementRef);
        console.log('delta', delta);
        console.log('onResizeStop, execute fit');
        // containerTerminal.onTerminalResize()
        const zz = terminal.mustGetAddon<FitAddon>('fit');
        const dim = zz.proposeDimensions();
        if (!dim) {
          console.log('dim is empty, ignore resize');
          return;
        }
        terminal.getTerminal().resize(dim.cols, dim.rows);
      }}
    >
      <div ref={terminalContainerRef} style={{ height: '100%' }}></div>
    </Resizable>
  );
};

export default ResizableTtyd;
