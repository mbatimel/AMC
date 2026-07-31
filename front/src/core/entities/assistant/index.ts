export type { AssistantAuthor, AssistantMessage } from './lib/types';

export { useAssistant } from './lib/useAssistant';
export { ASSISTANT_GREETING, ASSISTANT_SUGGESTIONS } from './lib/scripts';
export {
  $assistantHistory,
  $isAssistantOpen,
  assistantMessageSent,
  assistantOpened,
  assistantToggled,
} from './model';
