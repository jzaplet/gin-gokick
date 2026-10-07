import { RuleConfigSeverity, type UserConfig } from '@commitlint/types';

export default {
    extends: ['@commitlint/config-conventional'],
    rules: {
        'subject-case': [RuleConfigSeverity.Disabled],
        'body-max-line-length': [RuleConfigSeverity.Disabled],
        'footer-max-line-length': [RuleConfigSeverity.Disabled],
    },
} satisfies UserConfig;
