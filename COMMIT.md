# Pull Request Template

## Description

This PR enhances the Anthropic integration to make it more aligned with the OpenAI implementation. The changes include adding proper streaming support, implementing multimodal capabilities, improving error handling, and ensuring consistent behavior between the two LLM implementations.

Specifically, this PR addresses:

1. Implemented enhanced `stream` method to correctly handle streaming responses and properly accumulate text content
2. Added support for stop sequences and the `tool_choice` parameter in the `buildChatCompletionRequest` method
3. Improved error handling by wrapping errors with `ErrAnthropicChat` for consistent error reporting
4. Removed debug print statements for cleaner production code
5. Added multimodal support by properly handling image content in the `threadToChatCompletionMessages` function

Fixes #225

## Type of change

- [ ] Bug fix (non-breaking change which fixes an issue)
- [x] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] This change requires a documentation update

## How Has This Been Tested?

The changes have been tested using the example code provided in the repository:

- [x] Streaming example (`examples/llm/antropic/stream/main.go`) - Verified that streaming responses are correctly displayed
- [x] Multimodal example (`examples/llm/antropic/multimodal/main.go`) - Verified that image content is correctly processed and described

## Checklist:

- [x] I have read the [Contributing documentation](https://github.com/henomis/lingoose/blob/main/CONTRIBUTING.md).
- [x] I have read the [Code of conduct documentation](https://github.com/henomis/lingoose/blob/main/CODE_OF_CONDUCT.md).
- [x] I have performed a self-review of my own code
- [x] I have commented my code, particularly in hard-to-understand areas
- [x] I have made corresponding changes to the documentation
- [x] My changes generate no new warnings
- [x] I have added tests that prove my fix is effective or that my feature works
- [x] New and existing unit tests pass locally with my changes
- [x] I have checked my code and corrected any misspellings
