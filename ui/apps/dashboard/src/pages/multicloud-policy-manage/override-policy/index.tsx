import i18nInstance from '@/utils/i18n';
import { useState } from 'react';
import Panel from '@/components/panel';
import {
    Input,
    Button,
    Segmented,
    TableColumnProps,
    Tag,
    Space,
    Table,
    message,
    Popconfirm,
} from 'antd';
import { Icons } from '@/components/icons';
import { useQuery } from '@tanstack/react-query';
import {
    DeletePropagationPolicy,
} from '@/services/propagationpolicy.ts';
import { stringify } from 'yaml';
import { GetResource } from '@/services/unstructured.ts';
import {GetOverridePolicies, OverridePolicy} from "@/services/overridepolicy.ts";
export type PolicyScope = 'namespace-scope' | 'cluster-scope';
const OverridePolicyManage = () => {
    const [policyScope, setPolicyScope] =
        useState<PolicyScope>('namespace-scope');
    const { data, isLoading, refetch } = useQuery({
        queryKey: ['GetOverridePolicies', policyScope],
        queryFn: async () => {
            if (policyScope === 'cluster-scope')
                return {
                    overridepolicys: [],
                };
            const ret = await GetOverridePolicies();
            return ret.data || {};
        },
    });
    /*
    const [editorDrawerData, setEditorDrawerData] = useState<
        Omit<
            PropagationPolicyEditorDrawerProps,
            'onClose' | 'onUpdate' | 'onCreate'
        >
    >({
        open: false,
        mode: 'detail',
        name: '',
        namespace: '',
        propagationContent: '',
    });
    */
    const columns: TableColumnProps<OverridePolicy>[] = [
        {
            title: '命名空间',
            key: 'namespaceName',
            width: 200,
            render: (_, r) => {
                return r.objectMeta.namespace;
            },
        },
        {
            title: '策略名称',
            key: 'policyName',
            width: 200,
            render: (_, r) => {
                return r.objectMeta.name;
            },
        },
        {
            title: '调度器名称',
            key: 'schedulerName',
            dataIndex: 'schedulerName',
            width: 200,
        },
        {
            title: '关联集群',
            key: 'cluster',
            render: (_, r) => {
                if (!r?.clusterAffinity?.clusterNames) {
                    return '-';
                }
                return (
                    <div>
                        {r.clusterAffinity.clusterNames.map((key) => (
                            <Tag key={`${r.objectMeta.name}-${key}`}>{key}</Tag>
                        ))}
                    </div>
                );
            },
        },
        {
            title: '操作',
            key: 'op',
            width: 200,
            render: (_, r) => {
                return (
                    <Space.Compact>
                        <Button
                            size={'small'}
                            type="link"
                            onClick={async () => {
                                const ret = await GetResource({
                                    name: r.objectMeta.name,
                                    namespace: r.objectMeta.namespace,
                                    kind: 'propagationpolicy',
                                });
                                const content = stringify(ret.data);
                                /*
                                setEditorDrawerData({
                                    open: true,
                                    mode: 'detail',
                                    name: r.objectMeta.name,
                                    namespace: r.objectMeta.namespace,
                                    propagationContent: content,
                                });
                                */
                            }}
                        >
                            {i18nInstance.t('607e7a4f377fa66b0b28ce318aab841f')}
                        </Button>
                        <Button
                            size={'small'}
                            type="link"
                            onClick={async () => {
                                const ret = await GetResource({
                                    name: r.objectMeta.name,
                                    namespace: r.objectMeta.namespace,
                                    kind: 'propagationpolicy',
                                });
                                const content = stringify(ret.data);
                                /*
                                setEditorDrawerData({
                                    open: true,
                                    mode: 'edit',
                                    name: r.objectMeta.name,
                                    namespace: r.objectMeta.namespace,
                                    propagationContent: content,
                                });*/
                            }}
                        >
                            {i18nInstance.t('95b351c86267f3aedf89520959bce689')}
                        </Button>
                        <Popconfirm
                            placement="topRight"
                            title={`${i18nInstance.t('fc763fd5ddf637fe4ba1ac59e10b8d3a', '确认要删除')}${r.objectMeta.name}${i18nInstance.t('aa141bcb65729912b79cb27995a8989b', '调度策略么')}`}
                            onConfirm={async () => {
                                const ret = await DeletePropagationPolicy({
                                    isClusterScope: policyScope === 'cluster-scope',
                                    namespace: r.objectMeta.namespace,
                                    name: r.objectMeta.name,
                                });
                                if (ret.code === 200) {
                                    await messageApi.success(
                                        i18nInstance.t('0007d170de017dafc266aa03926d7f00'),
                                    );
                                    await refetch();
                                } else {
                                    await messageApi.error(
                                        i18nInstance.t('acf0664a54dc58d9d0377bb56e162092'),
                                    );
                                }
                            }}
                            okText={i18nInstance.t('e83a256e4f5bb4ff8b3d804b5473217a')}
                            cancelText={i18nInstance.t('625fb26b4b3340f7872b411f401e754c')}
                        >
                            <Button size={'small'} type="link" danger>
                                {i18nInstance.t('2f4aaddde33c9b93c36fd2503f3d122b')}
                            </Button>
                        </Popconfirm>
                    </Space.Compact>
                );
            },
        },
    ];
    const [messageApi, messageContextHolder] = message.useMessage();
    function resetEditorDrawerData() {
        /*
        setEditorDrawerData({
            open: false,
            mode: 'detail',
            name: '',
            namespace: '',
            propagationContent: '',
        });
        */
    }
    return (
        <Panel>
            <Segmented
                value={policyScope}
                style={{
                    marginBottom: 8,
                }}
                onChange={(value) => setPolicyScope(value as PolicyScope)}
                options={[
                    {
                        label: i18nInstance.t('bf15e71b2553d369585ace795d15ac3b'),
                        value: 'namespace-scope',
                    },
                    {
                        label: i18nInstance.t('860f29d8fc7a68113902db52885111d4'),
                        value: 'cluster-scope',
                    },
                ]}
            />

            <div className={'flex flex-row justify-between mb-4'}>
                <Input.Search
                    placeholder={i18nInstance.t('cfaff3e369b9bd51504feb59bf0972a0')}
                    className={'w-[400px]'}
                />
                <Button
                    type={'primary'}
                    icon={<Icons.add width={16} height={16} />}
                    className="flex flex-row items-center"
                    onClick={() => {
                        /*
                        setEditorDrawerData({
                            open: true,
                            mode: 'create',
                        });
                        */
                    }}
                >
                    {policyScope === 'namespace-scope'
                        ? i18nInstance.t('5ac6560da4f54522d590c5f8e939691b')
                        : i18nInstance.t('929e0cda9f7fdc960dafe6ef742ab088')}
                </Button>
            </div>
            <Table
                rowKey={(r: OverridePolicy) => r.objectMeta.name || ''}
                columns={columns}
                loading={isLoading}
                dataSource={data?.overridepolicys || []}
            />
            {/*
            <PropagationPolicyEditorDrawer
                open={editorDrawerData.open}
                name={editorDrawerData.name}
                namespace={editorDrawerData.namespace}
                mode={editorDrawerData.mode}
                propagationContent={editorDrawerData.propagationContent}
                onClose={() => {
                    setEditorDrawerData({
                        open: false,
                        mode: 'detail',
                        name: '',
                        namespace: '',
                        propagationContent: '',
                    });
                }}
                onCreate={async (ret) => {
                    if (ret.code === 200) {
                        await messageApi.success(
                            `${i18nInstance.t('8233550b23ab7acc2a9c3b2623c371dd', '新增调度策略成功')}`,
                        );
                        resetEditorDrawerData();
                        await refetch();
                    } else {
                        await messageApi.error(
                            `${i18nInstance.t('40eae6f51d50abb0f0132d7638682093', '新增调度策略失败')}`,
                        );
                    }
                }}
                onUpdate={async (ret) => {
                    if (ret.code === 200) {
                        await messageApi.success(
                            `${i18nInstance.t('f2224910b0d022374967254002eb756f', '编辑调度策略成功')}`,
                        );
                        resetEditorDrawerData();
                        await refetch();
                    } else {
                        await messageApi.error(
                            `${i18nInstance.t('5863fd1d291adf46d804f5801a79d0e1', '编辑调度策略失败')}`,
                        );
                    }
                }}
            />
            */}
            {messageContextHolder}
        </Panel>
    );
};
export default OverridePolicyManage;
