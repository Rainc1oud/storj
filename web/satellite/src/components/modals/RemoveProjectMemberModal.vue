// Copyright (C) 2023 Storj Labs, Inc.
// See LICENSE for copying information.

<template>
    <VModal :on-close="closeModal">
        <template #content>
            <div class="modal">
                <div class="modal__header">
                    <TeamMemberIcon />
                    <h1 class="modal__header__title">Remove Member</h1>
                </div>
                <p class="modal__info">
                    The following team members will be removed. This action cannot be undone.
                </p>
                <div class="modal__pm-container">
                    <div v-for="(member, key) in firstThreeSelected" :key="key" class="modal__project-member">
                        {{ member }}
                    </div>
                    <div v-if="selectedMembersLength > 3" class="modal__project-member">
                        + {{ selectedMembersLength - 3 }} more
                    </div>
                    <div class="modal__notice">
                        <strong>Please note:</strong> any access grants they have created will still provide them with full access. If necessary, please revoke these access grants to ensure the security of your data.
                    </div>
                </div>
                <div class="modal__button-container">
                    <VButton
                        label="Cancel"
                        width="100%"
                        height="48px"
                        font-size="14px"
                        border-radius="10px"
                        :on-press="closeModal"
                        :is-transparent="true"
                    />
                    <VButton
                        label="Remove"
                        :is-solid-delete="true"
                        icon="trash"
                        width="100%"
                        height="48px"
                        font-size="14px"
                        border-radius="10px"
                        :on-press="onRemove"
                    />
                </div>
            </div>
        </template>
    </VModal>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRouter } from 'vue-router';

import { RouteConfig } from '@/router';
import { AnalyticsHttpApi } from '@/api/analytics';
import { Project } from '@/types/projects';
import { AnalyticsErrorEventSource } from '@/utils/constants/analyticsEventNames';
import { useNotify } from '@/utils/hooks';
import { useAppStore } from '@/store/modules/appStore';
import { useProjectsStore } from '@/store/modules/projectsStore';
import { useProjectMembersStore } from '@/store/modules/projectMembersStore';
import { useConfigStore } from '@/store/modules/configStore';

import VModal from '@/components/common/VModal.vue';
import VButton from '@/components/common/VButton.vue';

import TeamMemberIcon from '@/../static/images/team/teamMember.svg';

const FIRST_PAGE = 1;
const configStore = useConfigStore();
const appStore = useAppStore();
const projectsStore = useProjectsStore();
const pmStore = useProjectMembersStore();
const notify = useNotify();
const router = useRouter();

const analytics: AnalyticsHttpApi = new AnalyticsHttpApi();

const firstThreeSelected = computed((): string[] => {
    return pmStore.state.selectedProjectMembersEmails.slice(0, 3);
});

const selectedMembersLength = computed((): number => {
    return pmStore.state.selectedProjectMembersEmails.length;
});

async function setProjectState(): Promise<void> {
    const projects: Project[] = await projectsStore.getProjects();
    if (!projects.length) {
        const onboardingPath = RouteConfig.OnboardingTour.with(configStore.firstOnboardingStep).path;

        analytics.pageVisit(onboardingPath);
        router.push(onboardingPath);

        return;
    }

    if (!projects.includes(projectsStore.state.selectedProject)) {
        projectsStore.selectProject(projects[0].id);
    }

    await pmStore.getProjectMembers(FIRST_PAGE, projectsStore.state.selectedProject.id);
}

async function onRemove(): Promise<void> {
    try {
        await pmStore.deleteProjectMembers(projectsStore.state.selectedProject.id);
        await setProjectState();
    } catch (error) {
        notify.error(`Error while deleting users from projectMembers. ${error.message}`, AnalyticsErrorEventSource.PROJECT_MEMBERS_HEADER);
        return;
    }

    notify.success('Members were successfully removed from project');
    pmStore.setSearchQuery('');
    closeModal();
}

/**
 * Closes remove team member modal.
 */
function closeModal(): void {
    appStore.removeActiveModal();
}
</script>

<style scoped lang="scss">
.modal {
    padding: 32px;
    font-family: 'font_regular', sans-serif;
    display: flex;
    flex-direction: column;
    max-width: 350px;

    @media screen and (width <= 615px) {
        padding: 30px 20px;
    }

    &__blur {
        position: absolute;
        top: 0;
        left: 0;
        height: 100%;
        width: 100%;
        background-color: rgb(229 229 229 / 20%);
        border-radius: 8px;
        z-index: 100;

        &__loader {
            width: 25px;
            height: 25px;
            position: absolute;
            right: 40px;
            top: 40px;
        }
    }

    :deep(.label-container) {
        margin-bottom: 8px;
    }

    :deep(.label-container__main__label) {
        font-family: 'font_bold', sans-serif;
        font-size: 14px;
        color: #56606d;
    }

    &__header {
        display: flex;
        align-items: center;
        padding-bottom: 16px;
        border-bottom: 1px solid var(--c-grey-2);

        &__title {
            font-family: 'font_bold', sans-serif;
            font-size: 24px;
            line-height: 31px;
            color: var(--c-grey-8);
            margin-left: 16px;
            text-align: left;
        }
    }

    &__info {
        font-size: 14px;
        line-height: 20px;
        color: var(--c-blue-6);
        padding: 16px 0;
        border-bottom: 1px solid var(--c-grey-2);
        text-align: left;
    }

    &__pm-container {
        border-bottom: 1px solid var(--c-grey-2);
    }

    &__button-container {
        width: 100%;
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-top: 16px;
        column-gap: 20px;

        @media screen and (width <= 600px) {
            margin-top: 20px;
            column-gap: unset;
            row-gap: 8px;
            flex-direction: column-reverse;
        }
    }

    &__notice {
        background: #fec;
        border: 1px solid #ffd78a;
        box-shadow: 0 7px 20px rgb(0 0 0 / 15%);
        border-radius: 10px;
        padding: 16px;
        gap: 16px;
        text-align: left;
        margin: 16px 0;

        strong {
            font-family: 'font_bold', sans-serif;
        }
    }

    &__project-member {
        width: fit-content;
        text-align: left;
        background: #f4f5f7;
        border-radius: 30px;
        padding: 7px 20px;
        gap: 10px;
        margin: 16px 0;
        font-family: 'font_medium', sans-serif;
    }
}

</style>
