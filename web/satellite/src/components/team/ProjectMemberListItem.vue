// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

<template>
    <table-item
        :class="{ 'owner': isProjectOwner }"
        :item="itemToRender"
        :selectable="true"
        :select-disabled="isProjectOwner"
        :selected="itemData.isSelected"
        :on-click="(_) => $emit('memberClick', itemData)"
        @selectClicked="($event) => $emit('selectClicked', $event)"
    />
</template>

<script setup lang="ts">
import { computed } from 'vue';

import { ProjectMember } from '@/types/projectMembers';
import { useResize } from '@/composables/resize';
import { useProjectsStore } from '@/store/modules/projectsStore';

import TableItem from '@/components/common/TableItem.vue';

const { isMobile, isTablet } = useResize();
const projectsStore = useProjectsStore();

const props = withDefaults(defineProps<{
    itemData: ProjectMember;
}>(), {
    itemData: () => new ProjectMember('', '', '', new Date(), ''),
});

const isProjectOwner = computed((): boolean => {
    return props.itemData.user.id === projectsStore.state.selectedProject.ownerId;
});

const itemToRender = computed((): { [key: string]: unknown | string[] } => {
    if (!isMobile.value && !isTablet.value) return { name: props.itemData.name, email: props.itemData.email, owner: isProjectOwner.value, date: props.itemData.localDate() };

    if (isTablet.value) {
        return { name: props.itemData.name, email: props.itemData.email, owner: isProjectOwner.value };
    }
    // TODO: change after adding actions button to list item
    return { name: props.itemData.name, email: props.itemData.email };
});
</script>

<style scoped lang="scss">
    :deep(.primary) {
        overflow: hidden;
        white-space: nowrap;
        text-overflow: ellipsis;
    }

    :deep(th) {
        max-width: 25rem;
    }

    @media screen and (width <= 940px) {

        :deep(th) {
            max-width: 10rem;
        }
    }
</style>
